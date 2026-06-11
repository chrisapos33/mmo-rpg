package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// LiveGitHubSource fetches real GitHub data via REST + GraphQL APIs.
// It caches the authenticated user's login/followers and each repo's primary language
// so RepoDependents can route to the correct package registry without re-fetching.
type LiveGitHubSource struct {
	token string

	mu           sync.Mutex
	login        string
	followers    int
	repoPrimLang map[string]string // "owner/name" → primary language (set after OwnedRepos)
}

func NewLiveGitHubSource(token string) *LiveGitHubSource {
	return &LiveGitHubSource{
		token:        token,
		repoPrimLang: make(map[string]string),
	}
}

// Username returns the authenticated user's GitHub login, caching after the first call.
func (s *LiveGitHubSource) Username(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.login != "" {
		return s.login, nil
	}
	var u struct {
		Login     string `json:"login"`
		Followers int    `json:"followers"`
	}
	if err := s.ghGet(ctx, "https://api.github.com/user", &u); err != nil {
		return "", err
	}
	s.login = u.Login
	s.followers = u.Followers
	return s.login, nil
}

// Followers returns the authenticated user's follower count, using the cached
// value from Username if it was already called.
func (s *LiveGitHubSource) Followers(ctx context.Context) (int, error) {
	s.mu.Lock()
	cached := s.followers
	s.mu.Unlock()
	if cached > 0 {
		return cached, nil
	}
	if _, err := s.Username(ctx); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.followers, nil
}

// OwnedRepos returns non-fork repos for the authenticated user, enriched with
// language breakdown, contributor count, CI/test detection, and star timestamps.
func (s *LiveGitHubSource) OwnedRepos(ctx context.Context) ([]RawRepo, error) {
	login, err := s.Username(ctx)
	if err != nil {
		return nil, err
	}

	type apiRepo struct {
		Name        string    `json:"name"`
		Fork        bool      `json:"fork"`
		Stars       int       `json:"stargazers_count"`
		Forks       int       `json:"forks_count"`
		OpenIssues  int       `json:"open_issues_count"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	var allRepos []apiRepo
	pageURL := "https://api.github.com/user/repos?per_page=100&type=owner&sort=pushed"
	for pageURL != "" {
		var page []apiRepo
		next, err := s.ghGetPaged(ctx, pageURL, &page)
		if err != nil {
			return nil, fmt.Errorf("repos page: %w", err)
		}
		allRepos = append(allRepos, page...)
		pageURL = next
	}

	var result []RawRepo
	for _, r := range allRepos {
		if r.Fork {
			continue
		}
		raw := RawRepo{
			Owner:      login,
			Name:       r.Name,
			Stars:      r.Stars,
			Forks:      r.Forks,
			OpenIssues: r.OpenIssues,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}

		var langs map[string]int64
		_ = s.ghGet(ctx, fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", login, r.Name), &langs)
		raw.Languages = langs

		if prim := primaryLanguage(langs); prim != "" {
			s.mu.Lock()
			s.repoPrimLang[login+"/"+r.Name] = prim
			s.mu.Unlock()
		}

		// Count external contributors from the first page (up to 100).
		type contributor struct {
			Login string `json:"login"`
		}
		var contribs []contributor
		_ = s.ghGet(ctx,
			fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors?per_page=100&anon=false", login, r.Name),
			&contribs,
		)
		for _, c := range contribs {
			if c.Login != login {
				raw.ExternalContributors++
			}
		}

		raw.HasCI = s.pathExists(ctx, login, r.Name, ".github/workflows") ||
			s.pathExists(ctx, login, r.Name, ".travis.yml") ||
			s.pathExists(ctx, login, r.Name, ".circleci/config.yml")

		raw.HasTests = s.pathExists(ctx, login, r.Name, "tests") ||
			s.pathExists(ctx, login, r.Name, "test") ||
			s.pathExists(ctx, login, r.Name, "spec")

		raw.StarredAt = s.fetchStarTimestamps(ctx, login, r.Name)

		result = append(result, raw)
	}
	return result, nil
}

// RepoDependents queries the package registry for the repo's primary language.
// Requires OwnedRepos to have been called first (to cache the primary language).
func (s *LiveGitHubSource) RepoDependents(ctx context.Context, owner, name string) (int, error) {
	s.mu.Lock()
	lang := s.repoPrimLang[owner+"/"+name]
	s.mu.Unlock()
	return FetchRegistryDependents(ctx, lang, name), nil
}

// ExternalMergedPRs returns merged PRs the user authored on repos they don't own.
// Paginates up to 300 results via the GitHub GraphQL search API.
func (s *LiveGitHubSource) ExternalMergedPRs(ctx context.Context) ([]RawPR, error) {
	login, err := s.Username(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
query($q: String!, $cursor: String) {
  search(query: $q, type: ISSUE, first: 50, after: $cursor) {
    nodes {
      ... on PullRequest {
        mergedAt
        repository { owner { login } name stargazerCount }
        reviewThreads { totalCount }
      }
    }
    pageInfo { hasNextPage endCursor }
  }
}`

	type searchData struct {
		Search struct {
			Nodes []struct {
				MergedAt   time.Time `json:"mergedAt"`
				Repository struct {
					Owner          struct{ Login string `json:"login"` } `json:"owner"`
					Name           string                                `json:"name"`
					StargazerCount int                                   `json:"stargazerCount"`
				} `json:"repository"`
				ReviewThreads struct {
					TotalCount int `json:"totalCount"`
				} `json:"reviewThreads"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"search"`
	}

	searchQ := fmt.Sprintf("is:pr is:merged author:%s -user:%s", login, login)
	var cursor *string
	var result []RawPR

	for len(result) < 300 {
		var data searchData
		if err := s.graphQL(ctx, q, map[string]any{"q": searchQ, "cursor": cursor}, &data); err != nil {
			return nil, fmt.Errorf("external PRs: %w", err)
		}
		for _, n := range data.Search.Nodes {
			result = append(result, RawPR{
				RepoOwner:         n.Repository.Owner.Login,
				RepoName:          n.Repository.Name,
				RepoStars:         n.Repository.StargazerCount,
				ReviewThreadCount: n.ReviewThreads.TotalCount,
				MergedAt:          n.MergedAt,
			})
		}
		if !data.Search.PageInfo.HasNextPage {
			break
		}
		c := data.Search.PageInfo.EndCursor
		cursor = &c
	}
	return result, nil
}

// ReviewsGiven returns code reviews submitted by the user on others' PRs,
// covering the 548-day scoring window via two overlapping 365-day GraphQL queries.
func (s *LiveGitHubSource) ReviewsGiven(ctx context.Context) ([]RawReview, error) {
	login, err := s.Username(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
query($from: DateTime!, $to: DateTime!) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      pullRequestReviewContributions(first: 100) {
        nodes {
          occurredAt
          pullRequestReview {
            pullRequest {
              repository { owner { login } name }
              author { login }
            }
          }
        }
      }
    }
  }
}`

	type reviewData struct {
		Viewer struct {
			ContributionsCollection struct {
				PullRequestReviewContributions struct {
					Nodes []struct {
						OccurredAt        time.Time `json:"occurredAt"`
						PullRequestReview struct {
							PullRequest struct {
								Repository struct {
									Owner struct{ Login string `json:"login"` } `json:"owner"`
									Name  string                                `json:"name"`
								} `json:"repository"`
								Author struct{ Login string `json:"login"` } `json:"author"`
							} `json:"pullRequest"`
						} `json:"pullRequestReview"`
					} `json:"nodes"`
				} `json:"pullRequestReviewContributions"`
			} `json:"contributionsCollection"`
		} `json:"viewer"`
	}

	now := time.Now()
	windows := twoWindows(now)
	seen := make(map[string]bool)
	var result []RawReview

	for _, w := range windows {
		vars := map[string]any{
			"from": w[0].UTC().Format(time.RFC3339),
			"to":   w[1].UTC().Format(time.RFC3339),
		}
		var data reviewData
		if err := s.graphQL(ctx, q, vars, &data); err != nil {
			return nil, fmt.Errorf("reviews given: %w", err)
		}
		for _, n := range data.Viewer.ContributionsCollection.PullRequestReviewContributions.Nodes {
			pr := n.PullRequestReview.PullRequest
			if pr.Author.Login == login {
				continue
			}
			key := pr.Repository.Owner.Login + "/" + pr.Repository.Name + "/" + n.OccurredAt.Format("2006-01-02")
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, RawReview{
				RepoOwner: pr.Repository.Owner.Login,
				RepoName:  pr.Repository.Name,
				CreatedAt: n.OccurredAt,
			})
		}
	}
	return result, nil
}

// ContributionDays returns per-(repo, date) commit activity within the scoring window.
// Uses two overlapping 365-day GraphQL windows to cover the full 548-day range.
func (s *LiveGitHubSource) ContributionDays(ctx context.Context) ([]RawContributionDay, error) {
	login, err := s.Username(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
query($from: DateTime!, $to: DateTime!) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      commitContributionsByRepository(maxRepositories: 100) {
        repository {
          owner { login }
          name
          stargazerCount
          isFork
        }
        contributions(first: 100) {
          nodes { occurredAt }
        }
      }
    }
  }
}`

	type contribData struct {
		Viewer struct {
			ContributionsCollection struct {
				CommitContributionsByRepository []struct {
					Repository struct {
						Owner          struct{ Login string `json:"login"` } `json:"owner"`
						Name           string                                `json:"name"`
						StargazerCount int                                   `json:"stargazerCount"`
						IsFork         bool                                  `json:"isFork"`
					} `json:"repository"`
					Contributions struct {
						Nodes []struct {
							OccurredAt time.Time `json:"occurredAt"`
						} `json:"nodes"`
					} `json:"contributions"`
				} `json:"commitContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"viewer"`
	}

	now := time.Now()
	windows := twoWindows(now)
	seen := make(map[string]bool)
	var result []RawContributionDay

	for _, w := range windows {
		vars := map[string]any{
			"from": w[0].UTC().Format(time.RFC3339),
			"to":   w[1].UTC().Format(time.RFC3339),
		}
		var data contribData
		if err := s.graphQL(ctx, q, vars, &data); err != nil {
			return nil, fmt.Errorf("contribution days: %w", err)
		}
		for _, byRepo := range data.Viewer.ContributionsCollection.CommitContributionsByRepository {
			repo := byRepo.Repository
			isOwn := repo.Owner.Login == login && !repo.IsFork
			for _, n := range byRepo.Contributions.Nodes {
				key := repo.Owner.Login + "/" + repo.Name + "/" + n.OccurredAt.Format("2006-01-02")
				if seen[key] {
					continue
				}
				seen[key] = true
				result = append(result, RawContributionDay{
					RepoOwner: repo.Owner.Login,
					RepoName:  repo.Name,
					IsOwnRepo: isOwn,
					Date:      n.OccurredAt,
					RepoStars: repo.StargazerCount,
					// For own repos, Ingest overwrites stars/contribs/dependents from
					// the fetched repo map. For external repos this is the source of truth.
					RepoHasExternalContribs: !isOwn,
				})
			}
		}
	}
	return result, nil
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func (s *LiveGitHubSource) ghGet(ctx context.Context, rawURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("github GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github GET %s returned %d", rawURL, resp.StatusCode)
	}
	return json.Unmarshal(body, dest)
}

// ghGetPaged performs a GET and returns the next-page URL parsed from the Link header.
func (s *LiveGitHubSource) ghGetPaged(ctx context.Context, rawURL string, dest any) (nextURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github GET %s returned %d", rawURL, resp.StatusCode)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return "", err
	}
	return parseLinkNext(resp.Header.Get("Link")), nil
}

func (s *LiveGitHubSource) graphQL(ctx context.Context, query string, variables map[string]any, dest any) error {
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": variables})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("graphql: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding graphql response: %w", err)
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}
	return json.Unmarshal(result.Data, dest)
}

// pathExists reports whether a file or directory exists in the repo via the Contents API.
func (s *LiveGitHubSource) pathExists(ctx context.Context, owner, repo, path string) bool {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// fetchStarTimestamps returns the timestamps from the first 100 stargazers,
// using the starred media type. Reserved for future star-velocity analysis.
func (s *LiveGitHubSource) fetchStarTimestamps(ctx context.Context, owner, repo string) []time.Time {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/stargazers?per_page=100", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github.star+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var stars []struct {
		StarredAt time.Time `json:"starred_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stars); err != nil {
		return nil
	}
	out := make([]time.Time, len(stars))
	for i, sg := range stars {
		out[i] = sg.StarredAt
	}
	return out
}

// twoWindows returns the two overlapping 365-day date ranges that together
// cover the full 548-day scoring window. GitHub's contributionsCollection
// has a hard 1-year maximum range per query.
func twoWindows(now time.Time) [2][2]time.Time {
	return [2][2]time.Time{
		{now.AddDate(0, 0, -548), now.AddDate(0, 0, -183)},
		{now.AddDate(0, 0, -365), now},
	}
}

// parseLinkNext extracts the URL tagged rel="next" from a GitHub Link header.
func parseLinkNext(link string) string {
	for _, part := range strings.Split(link, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, `rel="next"`) {
			continue
		}
		idx := strings.Index(part, "<")
		end := strings.Index(part, ">")
		if idx >= 0 && end > idx {
			return part[idx+1 : end]
		}
	}
	return ""
}

// primaryLanguage returns the language with the most bytes in the breakdown map.
func primaryLanguage(langs map[string]int64) string {
	type kv struct {
		k string
		v int64
	}
	pairs := make([]kv, 0, len(langs))
	for k, v := range langs {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })
	if len(pairs) == 0 {
		return ""
	}
	return pairs[0].k
}
