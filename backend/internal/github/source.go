package github

import (
	"context"
	"time"
)

// GitHubSource abstracts the GitHub API for the ingestion layer.
// Two implementations exist: LiveGitHubSource (real REST + GraphQL) and
// MockGitHubSource (deterministic fixture data).
//
// Contract: the scoring package never imports this package. Ingest() is the only
// function that crosses the boundary, translating raw GitHub data into scoring types.
type GitHubSource interface {
	// Username returns the authenticated user's GitHub login.
	Username(ctx context.Context) (string, error)

	// Followers returns the user's GitHub follower count.
	Followers(ctx context.Context) (int, error)

	// OwnedRepos returns repositories the user owns (not forks), with full metadata.
	// DependentCount is populated separately by Ingest via RepoDependents.
	OwnedRepos(ctx context.Context) ([]RawRepo, error)

	// ExternalMergedPRs returns merged pull requests to repos the user does NOT own.
	ExternalMergedPRs(ctx context.Context) ([]RawPR, error)

	// ReviewsGiven returns code reviews the user submitted on others' pull requests.
	ReviewsGiven(ctx context.Context) ([]RawReview, error)

	// ContributionDays returns per-(repo, date) commit activity within the scoring window.
	ContributionDays(ctx context.Context) ([]RawContributionDay, error)

	// RepoDependents returns the number of packages that depend on the given repo,
	// queried from the relevant package registry (npm, PyPI, crates.io).
	// Returns 0 if the repo is not a published package or the registry is unreachable.
	RepoDependents(ctx context.Context, owner, repo string) (int, error)
}

// RawRepo is raw repository data before mapping to scoring types.
type RawRepo struct {
	Owner                string
	Name                 string
	Stars                int
	Forks                int
	OpenIssues           int
	Languages            map[string]int64 // language → bytes of code
	ExternalContributors int
	DependentCount       int // populated by Ingest via RepoDependents after OwnedRepos returns
	HasCI                bool
	HasTests             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// StarredAt holds timestamped stargazer events (via Accept: application/vnd.github.star+json).
	// Not currently mapped to scoring.GitHubInput — reserved for future star-velocity analysis.
	StarredAt []time.Time
}

// RawPR is a merged pull request to a repo the user does NOT own.
type RawPR struct {
	RepoOwner         string
	RepoName          string
	RepoStars         int
	ReviewThreadCount int // reviews authored by others on this PR (review depth signal)
	MergedAt          time.Time
}

// RawReview is a code review the user gave on someone else's pull request.
type RawReview struct {
	RepoOwner string
	RepoName  string
	CreatedAt time.Time
}

// RawContributionDay is one day of commit activity on a specific repository.
type RawContributionDay struct {
	RepoOwner               string
	RepoName                string
	IsOwnRepo               bool
	Date                    time.Time
	RepoStars               int
	RepoHasExternalContribs bool
	RepoDependentCount      int
}
