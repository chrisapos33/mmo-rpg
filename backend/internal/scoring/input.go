package scoring

import "time"

// GitHubInput is the raw fetched data fed into the scoring engine.
// Populated by the GitHub ingestion layer; the engine itself has no DB or HTTP coupling.
type GitHubInput struct {
	Username  string
	Followers int
	FetchedAt time.Time // treated as "now" for recency decay; use time.Now() if zero

	// OwnedRepos: repositories the user created (not forks).
	OwnedRepos []OwnedRepo

	// ExternalPRs: PRs the user authored against repos they do NOT own, that were merged.
	// A maintainer merged them — this is the Collaboration gold signal.
	ExternalPRs []ExternalPR

	// ReviewsGiven: code reviews the user submitted on other people's PRs.
	ReviewsGiven []ReviewGiven

	// ActiveDays: one entry per (repo, calendar-day) on which the user pushed commits.
	// Repo-level context is needed so quality multipliers can be applied per day.
	ActiveDays []ActiveDay
}

// OwnedRepo is a repository the user created (IsFork == false in the GitHub API).
type OwnedRepo struct {
	Owner string
	Name  string

	Stars      int
	Forks      int
	OpenIssues int

	// HasTests and HasCI are detected by the ingestion layer via file-path conventions.
	HasTests bool // test files present (e.g. *_test.go, test_*.py, *.spec.ts)
	HasCI    bool // CI config present (.github/workflows/, .travis.yml, circle.yml, etc.)

	// Languages maps language name → bytes of code in that language.
	Languages map[string]int64

	// ExternalContributors is the count of distinct contributors who are not the repo owner.
	ExternalContributors int

	// DependentCount is the count of downstream packages that import this repo's code,
	// fetched from package registries (npm, PyPI, crates.io). Zero means unknown.
	DependentCount int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ExternalPR is a PR the user authored on a repo they do NOT own, that was merged by a maintainer.
type ExternalPR struct {
	RepoOwner string
	RepoName  string
	RepoStars int

	// ReviewThreadCount is the number of review comments/threads the PR received before merge.
	// Hard to fake: a maintainer and reviewers spent time on this.
	ReviewThreadCount int

	MergedAt time.Time
}

// ReviewGiven is a code review the user submitted on another person's PR.
type ReviewGiven struct {
	RepoOwner string
	RepoName  string
	CreatedAt time.Time
}

// ActiveDay is a distinct calendar day on which the user pushed commits to a specific repo.
type ActiveDay struct {
	RepoOwner string
	RepoName  string
	IsOwnRepo bool
	Date      time.Time

	// Repo-level signals used for validation and confidence weighting.
	RepoStars               int
	RepoHasExternalContribs bool
	RepoDependentCount      int
}
