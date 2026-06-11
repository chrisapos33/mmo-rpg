package github_test

import (
	"context"
	"testing"
	"time"

	gh "github.com/chrisapos3/mmo-rpg/internal/github"
)

func TestIngest_FieldMapping(t *testing.T) {
	input, err := gh.Ingest(context.Background(), gh.NewMockGitHubSource())
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	if input.Username != "mockdev" {
		t.Errorf("Username = %q, want mockdev", input.Username)
	}
	if input.Followers != 94 {
		t.Errorf("Followers = %d, want 94", input.Followers)
	}
	if got, want := len(input.OwnedRepos), 2; got != want {
		t.Errorf("OwnedRepos count = %d, want %d", got, want)
	}
	if got, want := len(input.ExternalPRs), 3; got != want {
		t.Errorf("ExternalPRs count = %d, want %d", got, want)
	}
	if got, want := len(input.ReviewsGiven), 12; got != want {
		t.Errorf("ReviewsGiven count = %d, want %d", got, want)
	}
	// 120 own-repo + 40 external = 160 active days
	if got, want := len(input.ActiveDays), 160; got != want {
		t.Errorf("ActiveDays count = %d, want %d", got, want)
	}
}

func TestIngest_DependentsPopulated(t *testing.T) {
	input, err := gh.Ingest(context.Background(), gh.NewMockGitHubSource())
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	for _, r := range input.OwnedRepos {
		if r.Name == "fast-cache" {
			if r.DependentCount != 200 {
				t.Errorf("fast-cache DependentCount = %d, want 200 (from RepoDependents)", r.DependentCount)
			}
			return
		}
	}
	t.Fatal("fast-cache repo not found in OwnedRepos")
}

func TestIngest_OwnRepoEnrichment(t *testing.T) {
	input, err := gh.Ingest(context.Background(), gh.NewMockGitHubSource())
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	for _, d := range input.ActiveDays {
		if !d.IsOwnRepo || d.RepoName != "fast-cache" {
			continue
		}
		// ContributionDays sets RepoDependentCount=0 for own-repo days;
		// Ingest must overwrite it from the repo map (RepoDependents returns 200).
		if d.RepoDependentCount != 200 {
			t.Errorf("own-repo fast-cache day: RepoDependentCount = %d, want 200 (enriched from repo map, not ContributionDays)", d.RepoDependentCount)
		}
		if !d.RepoHasExternalContribs {
			t.Errorf("own-repo fast-cache day: RepoHasExternalContribs = false, want true (7 ext contributors)")
		}
		return
	}
	t.Fatal("no own-repo active day for fast-cache found in ActiveDays")
}

// TestIngest_OwnRepoPRsFilteredOut is the linchpin test for Collaboration classification.
//
// The live source uses `-user:{login}` in its GraphQL search query, which is an
// approximation: it excludes personal-account repos but can leak org repos the user owns.
// Ingest adds a belt-and-suspenders filter: any PR whose RepoOwner == authenticated user
// is dropped before it reaches the scoring engine.
//
// This test wires a source that injects two own-repo PRs alongside two genuine external
// PRs and asserts that only the two external ones survive into scoring.GitHubInput.ExternalPRs.
func TestIngest_OwnRepoPRsFilteredOut(t *testing.T) {
	input, err := gh.Ingest(context.Background(), &mixedPRSource{})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}

	if got, want := len(input.ExternalPRs), 2; got != want {
		t.Fatalf("ExternalPRs count = %d, want %d (2 own-repo PRs must be dropped)", got, want)
	}
	for _, pr := range input.ExternalPRs {
		if pr.RepoOwner == "testuser" {
			t.Errorf("own-repo PR for testuser/%s was not filtered out", pr.RepoName)
		}
	}
}

// mixedPRSource returns a user "testuser" with:
//   - 2 own-repo PRs (RepoOwner == "testuser") — must be filtered by Ingest
//   - 2 external PRs (RepoOwner == "someorg") — must pass through to ExternalPRs
type mixedPRSource struct{}

func (s *mixedPRSource) Username(_ context.Context) (string, error)  { return "testuser", nil }
func (s *mixedPRSource) Followers(_ context.Context) (int, error)    { return 0, nil }
func (s *mixedPRSource) OwnedRepos(_ context.Context) ([]gh.RawRepo, error) {
	return []gh.RawRepo{
		{Owner: "testuser", Name: "my-lib", Stars: 50, Languages: map[string]int64{"Go": 10000}},
		{Owner: "testuser", Name: "my-cli", Stars: 5},
	}, nil
}
func (s *mixedPRSource) RepoDependents(_ context.Context, _, _ string) (int, error) { return 0, nil }
func (s *mixedPRSource) ExternalMergedPRs(_ context.Context) ([]gh.RawPR, error) {
	now := time.Now()
	return []gh.RawPR{
		// Own-repo PRs — RepoOwner matches Username(); Ingest must drop these.
		{RepoOwner: "testuser", RepoName: "my-lib", RepoStars: 50, ReviewThreadCount: 3, MergedAt: now.AddDate(0, 0, -10)},
		{RepoOwner: "testuser", RepoName: "my-cli", RepoStars: 5, ReviewThreadCount: 1, MergedAt: now.AddDate(0, 0, -20)},
		// External PRs — different owner; must survive into ExternalPRs.
		{RepoOwner: "someorg", RepoName: "alpha", RepoStars: 5000, ReviewThreadCount: 4, MergedAt: now.AddDate(0, 0, -30)},
		{RepoOwner: "someorg", RepoName: "beta", RepoStars: 2000, ReviewThreadCount: 2, MergedAt: now.AddDate(0, 0, -60)},
	}, nil
}
func (s *mixedPRSource) ReviewsGiven(_ context.Context) ([]gh.RawReview, error)           { return nil, nil }
func (s *mixedPRSource) ContributionDays(_ context.Context) ([]gh.RawContributionDay, error) {
	return nil, nil
}
