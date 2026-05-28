package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
)

// ExchangeCode exchanges an OAuth code for a GitHub access token.
func ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s — %s", result.Error, result.ErrorDesc)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("github returned empty access token")
	}
	return result.AccessToken, nil
}

// User represents the GitHub API user response.
type User struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	Name        *string `json:"name"`
	AvatarURL   string  `json:"avatar_url"`
	PublicRepos int     `json:"public_repos"`
	Followers   int     `json:"followers"`
}

// Repo represents a single GitHub repository.
type Repo struct {
	Name            string  `json:"name"`
	Language        *string `json:"language"`
	StargazersCount int     `json:"stargazers_count"`
	Fork            bool    `json:"fork"`
}

// Stats is the aggregated profile computed from user + repos.
type Stats struct {
	User         User
	TotalStars   int
	OriginalRepos int
	TopLanguages []string
	ContribScore int
}

// FetchUser returns the authenticated GitHub user.
func FetchUser(ctx context.Context, token string) (*User, error) {
	var u User
	if err := ghGet(ctx, token, "https://api.github.com/user", &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// FetchRepos returns up to 100 repos for the authenticated user.
func FetchRepos(ctx context.Context, token string) ([]Repo, error) {
	var repos []Repo
	err := ghGet(ctx, token,
		"https://api.github.com/user/repos?per_page=100&sort=pushed&type=owner",
		&repos,
	)
	return repos, err
}

// AggregateStats computes top languages, star count, and a contribution score.
func AggregateStats(user *User, repos []Repo) *Stats {
	langFreq := map[string]int{}
	totalStars := 0
	origRepos := 0

	for _, r := range repos {
		if r.Fork {
			continue
		}
		origRepos++
		totalStars += r.StargazersCount
		if r.Language != nil && *r.Language != "" {
			langFreq[*r.Language]++
		}
	}

	topLangs := topN(langFreq, 5)
	score := totalStars*5 + origRepos*10 + user.Followers*3
	if score > 1000 {
		score = 1000
	}

	return &Stats{
		User:          *user,
		TotalStars:    totalStars,
		OriginalRepos: origRepos,
		TopLanguages:  topLangs,
		ContribScore:  score,
	}
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func ghGet(ctx context.Context, token, rawURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("github api request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api %s returned %d: %s", rawURL, resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, dest)
}

func topN(freq map[string]int, n int) []string {
	type kv struct {
		k string
		v int
	}
	pairs := make([]kv, 0, len(freq))
	for k, v := range freq {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })

	out := make([]string, 0, n)
	for _, p := range pairs {
		if len(out) >= n {
			break
		}
		out = append(out, p.k)
	}
	return out
}

// AuthorizeURL builds the GitHub OAuth authorization URL.
func AuthorizeURL(clientID, redirectURI, state string) string {
	v := url.Values{
		"client_id":    {clientID},
		"redirect_uri": {redirectURI},
		"scope":        {"read:user public_repo"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + v.Encode()
}
