package github

import (
	"context"
	"time"
)

// MockGitHubSource returns deterministic fixture data for a mid-career developer.
// Activate via MOCK_GITHUB=true so the ingestion→scoring pipeline is testable
// without live API calls or a real GitHub account.
//
// Profile "mockdev":
//   - fast-cache: 480★, Go+Shell, CI+tests, 7 external contributors, 200 dependents
//   - cli-tools: 35★, Python, no CI/tests
//   - 3 merged PRs to hashicorp/terraform, grafana/grafana, prometheus/prometheus
//   - 12 reviews given across 3 repos over the past 300 days
//   - 120 own-repo contribution days on fast-cache + 40 external days on kubernetes/kubernetes
type MockGitHubSource struct{}

func NewMockGitHubSource() *MockGitHubSource { return &MockGitHubSource{} }

func (m *MockGitHubSource) Username(_ context.Context) (string, error) {
	return "mockdev", nil
}

func (m *MockGitHubSource) Followers(_ context.Context) (int, error) {
	return 94, nil
}

func (m *MockGitHubSource) OwnedRepos(_ context.Context) ([]RawRepo, error) {
	now := time.Now()
	return []RawRepo{
		{
			Owner:                "mockdev",
			Name:                 "fast-cache",
			Stars:                480,
			Forks:                42,
			OpenIssues:           8,
			Languages:            map[string]int64{"Go": 72000, "Shell": 3000},
			ExternalContributors: 7,
			HasCI:                true,
			HasTests:             true,
			CreatedAt:            now.AddDate(-2, -3, 0),
			UpdatedAt:            now.AddDate(0, 0, -12),
		},
		{
			Owner:                "mockdev",
			Name:                 "cli-tools",
			Stars:                35,
			Forks:                3,
			OpenIssues:           2,
			Languages:            map[string]int64{"Python": 18000},
			ExternalContributors: 0,
			HasCI:                false,
			HasTests:             false,
			CreatedAt:            now.AddDate(-1, -1, 0),
			UpdatedAt:            now.AddDate(0, 0, -45),
		},
	}, nil
}

func (m *MockGitHubSource) RepoDependents(_ context.Context, owner, name string) (int, error) {
	if owner == "mockdev" && name == "fast-cache" {
		return 200, nil
	}
	return 0, nil
}

func (m *MockGitHubSource) ExternalMergedPRs(_ context.Context) ([]RawPR, error) {
	now := time.Now()
	return []RawPR{
		{
			RepoOwner:         "hashicorp",
			RepoName:          "terraform",
			RepoStars:         42000,
			ReviewThreadCount: 5,
			MergedAt:          now.AddDate(0, 0, -45),
		},
		{
			RepoOwner:         "grafana",
			RepoName:          "grafana",
			RepoStars:         58000,
			ReviewThreadCount: 3,
			MergedAt:          now.AddDate(0, 0, -120),
		},
		{
			RepoOwner:         "prometheus",
			RepoName:          "prometheus",
			RepoStars:         53000,
			ReviewThreadCount: 7,
			MergedAt:          now.AddDate(0, 0, -200),
		},
	}, nil
}

func (m *MockGitHubSource) ReviewsGiven(_ context.Context) ([]RawReview, error) {
	now := time.Now()
	repos := []struct{ owner, name string }{
		{"hashicorp", "terraform"},
		{"grafana", "grafana"},
		{"kubernetes", "kubernetes"},
	}
	reviews := make([]RawReview, 12)
	for i := range reviews {
		r := repos[i%len(repos)]
		reviews[i] = RawReview{
			RepoOwner: r.owner,
			RepoName:  r.name,
			CreatedAt: now.AddDate(0, 0, -(i*25 + 10)),
		}
	}
	return reviews, nil
}

func (m *MockGitHubSource) ContributionDays(_ context.Context) ([]RawContributionDay, error) {
	now := time.Now()
	days := make([]RawContributionDay, 0, 160)

	// 120 own-repo days on fast-cache, roughly every 3 days over the past year.
	// RepoDependentCount is intentionally 0 — Ingest overwrites it from the repo map
	// (RepoDependents returns 200 for fast-cache), so the enrichment path is exercised.
	for i := 0; i < 120; i++ {
		days = append(days, RawContributionDay{
			RepoOwner:               "mockdev",
			RepoName:                "fast-cache",
			IsOwnRepo:               true,
			Date:                    now.AddDate(0, 0, -(i*3 + 1)),
			RepoStars:               480,
			RepoHasExternalContribs: true,
			RepoDependentCount:      0,
		})
	}

	// 40 external-repo days on kubernetes/kubernetes.
	for i := 0; i < 40; i++ {
		days = append(days, RawContributionDay{
			RepoOwner:               "kubernetes",
			RepoName:                "kubernetes",
			IsOwnRepo:               false,
			Date:                    now.AddDate(0, 0, -(i*5 + 2)),
			RepoStars:               110000,
			RepoHasExternalContribs: true,
			RepoDependentCount:      0,
		})
	}

	return days, nil
}
