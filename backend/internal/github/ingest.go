package github

import (
	"context"
	"fmt"
	"time"

	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

// Ingest fetches all GitHub data via source and translates it into scoring.GitHubInput.
// This is the sole crossing point between the transport layer and the pure scoring engine —
// the engine never knows HTTP exists.
func Ingest(ctx context.Context, source GitHubSource) (scoring.GitHubInput, error) {
	username, err := source.Username(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("username: %w", err)
	}

	followers, err := source.Followers(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("followers: %w", err)
	}

	rawRepos, err := source.OwnedRepos(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("owned repos: %w", err)
	}

	// Populate DependentCount for each owned repo and build a lookup map for
	// enriching own-repo contribution days (GraphQL sources don't carry this).
	repoMap := make(map[string]RawRepo, len(rawRepos))
	for i, r := range rawRepos {
		deps, depErr := source.RepoDependents(ctx, r.Owner, r.Name)
		if depErr != nil {
			deps = 0
		}
		rawRepos[i].DependentCount = deps
		repoMap[r.Owner+"/"+r.Name] = rawRepos[i]
	}

	rawPRs, err := source.ExternalMergedPRs(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("external PRs: %w", err)
	}

	rawReviews, err := source.ReviewsGiven(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("reviews given: %w", err)
	}

	rawDays, err := source.ContributionDays(ctx)
	if err != nil {
		return scoring.GitHubInput{}, fmt.Errorf("contribution days: %w", err)
	}

	return buildScoringInput(username, followers, rawRepos, rawPRs, rawReviews, rawDays, repoMap), nil
}

func buildScoringInput(
	username string,
	followers int,
	rawRepos []RawRepo,
	rawPRs []RawPR,
	rawReviews []RawReview,
	rawDays []RawContributionDay,
	repoMap map[string]RawRepo,
) scoring.GitHubInput {
	// Defensive filter: the live source uses `-user:{login}` in its search query,
	// which is an approximation. Belt-and-suspenders: drop any PR whose repo owner
	// matches the authenticated user so own-repo PRs never reach the Collaboration engine.
	ownedRepos := make([]scoring.OwnedRepo, len(rawRepos))
	for i, r := range rawRepos {
		ownedRepos[i] = scoring.OwnedRepo{
			Owner:                r.Owner,
			Name:                 r.Name,
			Stars:                r.Stars,
			Forks:                r.Forks,
			OpenIssues:           r.OpenIssues,
			HasTests:             r.HasTests,
			HasCI:                r.HasCI,
			Languages:            r.Languages,
			ExternalContributors: r.ExternalContributors,
			DependentCount:       r.DependentCount,
			CreatedAt:            r.CreatedAt,
			UpdatedAt:            r.UpdatedAt,
		}
	}

	externalPRs := make([]scoring.ExternalPR, 0, len(rawPRs))
	for _, pr := range rawPRs {
		if pr.RepoOwner == username {
			continue // own-repo PR slipped through the source filter — drop it
		}
		externalPRs = append(externalPRs, scoring.ExternalPR{
			RepoOwner:         pr.RepoOwner,
			RepoName:          pr.RepoName,
			RepoStars:         pr.RepoStars,
			ReviewThreadCount: pr.ReviewThreadCount,
			MergedAt:          pr.MergedAt,
		})
	}

	reviewsGiven := make([]scoring.ReviewGiven, len(rawReviews))
	for i, rv := range rawReviews {
		reviewsGiven[i] = scoring.ReviewGiven{
			RepoOwner: rv.RepoOwner,
			RepoName:  rv.RepoName,
			CreatedAt: rv.CreatedAt,
		}
	}

	activeDays := make([]scoring.ActiveDay, 0, len(rawDays))
	for _, d := range rawDays {
		ad := scoring.ActiveDay{
			RepoOwner: d.RepoOwner,
			RepoName:  d.RepoName,
			IsOwnRepo: d.IsOwnRepo,
			Date:      d.Date,
		}
		if d.IsOwnRepo {
			// Enrich from the freshly-fetched repo map, which carries DependentCount
			// from RepoDependents. GraphQL/REST sources don't provide this per day.
			if r, ok := repoMap[d.RepoOwner+"/"+d.RepoName]; ok {
				ad.RepoStars = r.Stars
				ad.RepoHasExternalContribs = r.ExternalContributors > 0
				ad.RepoDependentCount = r.DependentCount
			}
		} else {
			ad.RepoStars = d.RepoStars
			ad.RepoHasExternalContribs = d.RepoHasExternalContribs
			ad.RepoDependentCount = d.RepoDependentCount
		}
		activeDays = append(activeDays, ad)
	}

	return scoring.GitHubInput{
		Username:     username,
		Followers:    followers,
		FetchedAt:    time.Now(),
		OwnedRepos:   ownedRepos,
		ExternalPRs:  externalPRs,
		ReviewsGiven: reviewsGiven,
		ActiveDays:   activeDays,
	}
}
