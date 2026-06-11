package scoring_test

import (
	"testing"
	"time"

	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

// now is a fixed reference point so all decay calculations are deterministic.
var now = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

// daysAgo returns a time exactly n days before now.
func daysAgo(n int) time.Time {
	return now.Add(-time.Duration(n) * 24 * time.Hour)
}

// ── Fixtures ────────────────────────────────────────────────────────────────

// zeroInput has no activity at all.
func zeroInput() scoring.GitHubInput {
	return scoring.GitHubInput{FetchedAt: now}
}

// soloDevInput: active developer but entirely self-contained — no external validators.
// All repos have zero stars, zero forks, zero external contributors, zero dependents.
// Expected: Output = 0 (no validated days), Collaboration = 0, Trust near-zero.
func soloDevInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "solodev",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:      "solodev",
				Name:       "private-tool",
				Stars:      0,
				HasTests:   false,
				HasCI:      false,
				Languages:  map[string]int64{"Go": 20000},
				CreatedAt:  daysAgo(400),
				UpdatedAt:  daysAgo(5),
			},
		},
		ActiveDays: activeDays("solodev", "private-tool", false, 0, false, 0, 200),
	}
}

// starredSoloDev: same solo dev but their repo has gathered stars. Output should be > 0.
func starredSoloDevInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "starreddev",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:      "starreddev",
				Name:       "cool-tool",
				Stars:      50,
				Forks:      0,
				HasTests:   false,
				HasCI:      false,
				Languages:  map[string]int64{"Python": 30000},
				CreatedAt:  daysAgo(500),
				UpdatedAt:  daysAgo(10),
			},
		},
		ActiveDays: activeDays("starreddev", "cool-tool", true, 50, false, 0, 150),
	}
}

// craftedDevInput: developer with tests, CI, long-lived repos, reviewed PRs.
// Expected: Craft score should be meaningfully higher than soloDevInput.
func craftedDevInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "craftsman",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:      "craftsman",
				Name:       "well-tested-lib",
				Stars:      80,
				Forks:      5,
				OpenIssues: 3,
				HasTests:   true,
				HasCI:      true,
				Languages:  map[string]int64{"Go": 50000, "Makefile": 5000},
				ExternalContributors: 3,
				CreatedAt:  daysAgo(800),
				UpdatedAt:  daysAgo(14),
			},
		},
		ExternalPRs: []scoring.ExternalPR{
			{RepoOwner: "popular-org", RepoName: "big-project", RepoStars: 2000, ReviewThreadCount: 4, MergedAt: daysAgo(60)},
			{RepoOwner: "another-org", RepoName: "framework", RepoStars: 500, ReviewThreadCount: 2, MergedAt: daysAgo(120)},
		},
		ActiveDays: activeDays("craftsman", "well-tested-lib", true, 80, true, 0, 180),
	}
}

// openSourceContribInput: heavy external contributor — the ideal Collaboration profile.
// Many merged PRs to popular repos, substantive review threads, diverse orgs.
func openSourceContribInput() scoring.GitHubInput {
	prs := []scoring.ExternalPR{
		{RepoOwner: "kubernetes", RepoName: "kubernetes", RepoStars: 110000, ReviewThreadCount: 8, MergedAt: daysAgo(30)},
		{RepoOwner: "golang", RepoName: "go", RepoStars: 125000, ReviewThreadCount: 5, MergedAt: daysAgo(90)},
		{RepoOwner: "microsoft", RepoName: "vscode", RepoStars: 165000, ReviewThreadCount: 3, MergedAt: daysAgo(150)},
		{RepoOwner: "facebook", RepoName: "react", RepoStars: 228000, ReviewThreadCount: 7, MergedAt: daysAgo(200)},
		{RepoOwner: "torvalds", RepoName: "linux", RepoStars: 180000, ReviewThreadCount: 12, MergedAt: daysAgo(250)},
		{RepoOwner: "rust-lang", RepoName: "rust", RepoStars: 98000, ReviewThreadCount: 6, MergedAt: daysAgo(300)},
	}
	reviews := make([]scoring.ReviewGiven, 40)
	for i := range reviews {
		reviews[i] = scoring.ReviewGiven{RepoOwner: "some-org", RepoName: "some-repo", CreatedAt: daysAgo(i * 5)}
	}

	return scoring.GitHubInput{
		Username:  "topcontrib",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:      "topcontrib",
				Name:       "mylib",
				Stars:      300,
				Forks:      40,
				OpenIssues: 10,
				HasTests:   true,
				HasCI:      true,
				Languages:  map[string]int64{"Go": 80000, "Python": 20000},
				ExternalContributors: 12,
				DependentCount: 50,
				CreatedAt:  daysAgo(1200),
				UpdatedAt:  daysAgo(7),
			},
		},
		ExternalPRs:  prs,
		ReviewsGiven: reviews,
		ActiveDays:   activeDays("topcontrib", "mylib", true, 300, true, 50, 300),
	}
}

// specialistInput: all work in one language on a validated repo. High concentration.
func specialistInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "rustacean",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:      "rustacean",
				Name:       "fast-parser",
				Stars:      120,
				Forks:      8,
				HasTests:   true,
				HasCI:      true,
				Languages:  map[string]int64{"Rust": 95000, "Shell": 2000},
				ExternalContributors: 4,
				CreatedAt:  daysAgo(600),
				UpdatedAt:  daysAgo(20),
			},
		},
		ActiveDays: activeDays("rustacean", "fast-parser", true, 120, true, 0, 200),
	}
}

// generalistInput: distributed work across many languages on validated repos.
func generalistInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "polyglot",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:     "polyglot",
				Name:      "web-backend",
				Stars:     60,
				Forks:     5,
				Languages: map[string]int64{"Go": 40000, "Python": 35000, "TypeScript": 30000, "Rust": 20000, "Haskell": 15000},
				ExternalContributors: 3,
				CreatedAt: daysAgo(700),
				UpdatedAt: daysAgo(15),
			},
		},
		ActiveDays: activeDays("polyglot", "web-backend", true, 60, true, 0, 150),
	}
}

// durableBuilderInput: author of a popular library created ~2 years ago with 5000 package
// dependents today, but almost no recent commit activity. The 5 active days all fall in
// the 300-500 day range — within the 18-month window, but the 180-day cadence half-life
// decays them to 17-32% of face value (Output ≈ p10). Dependents are a present-tense
// snapshot with NO decay, so Influence should stay at p99.
func durableBuilderInput() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "durablebuilder",
		FetchedAt: now,
		OwnedRepos: []scoring.OwnedRepo{
			{
				Owner:                "durablebuilder",
				Name:                 "popular-lib",
				Stars:                8000,
				Forks:                600,
				OpenIssues:           45,
				HasTests:             true,
				HasCI:                true,
				Languages:            map[string]int64{"Python": 80000, "Go": 20000},
				ExternalContributors: 30,
				DependentCount:       5000,
				CreatedAt:            daysAgo(730), // ~2 years old
				UpdatedAt:            daysAgo(250), // last commit ~8 months ago
			},
		},
		ActiveDays: activeDaysSpread("durablebuilder", "popular-lib", true, 8000, true, 5000, 300, 500, 5),
	}
}

// recencyInput: two developers with identical total activity, one recent and one old.
// The recent one should score higher on Output.
func recencyInputRecent() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "recent-dev",
		FetchedAt: now,
		ActiveDays: activeDaysSpread("recent-dev", "repo", true, 30, false, 0,
			0, 90, 50), // 50 days within last 90 days
	}
}

func recencyInputOld() scoring.GitHubInput {
	return scoring.GitHubInput{
		Username:  "old-dev",
		FetchedAt: now,
		ActiveDays: activeDaysSpread("old-dev", "repo", true, 30, false, 0,
			400, 490, 50), // 50 days from 400–490 days ago (outside window or heavily decayed)
	}
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestCompute_ZeroInput(t *testing.T) {
	s := scoring.Compute(zeroInput())

	assertZero(t, "Output.Percentile", s.Output.Percentile)
	assertZero(t, "Craft.Percentile", s.Craft.Percentile)
	assertZero(t, "Influence.Percentile", s.Influence.Percentile)
	assertZero(t, "Collaboration.Percentile", s.Collaboration.Percentile)
	assertZero(t, "Range.Percentile", s.Range.Percentile)
	assertZero(t, "Trust", s.Trust)
}

func TestCompute_UnvalidatedActivity_OutputIsZero(t *testing.T) {
	// soloDevInput has 200 active days but all on a repo with 0 stars / 0 external contribs.
	// The engine must ignore those days — activity in a vacuum signals nothing.
	// The repo IS maintained for 1+ year (longevity signal, confidence 0.52), but that
	// confidence is below the 0.70 Trust threshold — self-controlled signals never raise Trust.
	s := scoring.Compute(soloDevInput())

	assertZero(t, "Output.Raw (unvalidated days must not count)", s.Output.Raw)
	assertZero(t, "Output.Percentile", s.Output.Percentile)
	assertZero(t, "Collaboration.Raw", s.Collaboration.Raw)
	assertZero(t, "Trust (longevity conf 0.52 is below the 0.70 threshold)", s.Trust)
}

func TestCompute_StarredRepo_OutputAboveZero(t *testing.T) {
	s := scoring.Compute(starredSoloDevInput())

	if s.Output.Raw <= 0 {
		t.Errorf("Output.Raw = %v, want > 0 (starred repo should validate activity)", s.Output.Raw)
	}
	if s.Output.Percentile <= 0 {
		t.Errorf("Output.Percentile = %v, want > 0", s.Output.Percentile)
	}
}

func TestCompute_CraftSignals(t *testing.T) {
	uncraft := scoring.Compute(soloDevInput())
	crafted := scoring.Compute(craftedDevInput())

	if crafted.Craft.Raw <= uncraft.Craft.Raw {
		t.Errorf("craftedDev Craft.Raw %v should exceed solodev %v", crafted.Craft.Raw, uncraft.Craft.Raw)
	}
	if crafted.Craft.Percentile <= 0 {
		t.Errorf("Craft.Percentile = %v, want > 0", crafted.Craft.Percentile)
	}
	// Review depth should be reflected
	if crafted.Craft.Raw < 10 {
		t.Errorf("craftedDev has reviewed PRs and tests/CI — Craft.Raw = %v, expected >= 10", crafted.Craft.Raw)
	}
}

func TestCompute_Collaboration_ExternalPRsScore(t *testing.T) {
	noCollab := scoring.Compute(starredSoloDevInput())
	withCollab := scoring.Compute(craftedDevInput())
	heavyCollab := scoring.Compute(openSourceContribInput())

	if withCollab.Collaboration.Raw <= noCollab.Collaboration.Raw {
		t.Errorf("developer with external PRs should have higher Collaboration than one without")
	}
	if heavyCollab.Collaboration.Percentile <= withCollab.Collaboration.Percentile {
		t.Errorf("heavy open-source contributor should rank higher on Collaboration")
	}
	// Heavy contributor should be in a high percentile
	if heavyCollab.Collaboration.Percentile < 80 {
		t.Errorf("top OSS contributor Collaboration.Percentile = %.1f, want >= 80",
			heavyCollab.Collaboration.Percentile)
	}
}

func TestCompute_Trust_ExternalPRsRaiseTrust(t *testing.T) {
	solo := scoring.Compute(starredSoloDevInput())
	contrib := scoring.Compute(openSourceContribInput())

	// Stars (conf 0.55) and star-validated activity (conf 0.62) are below the 0.70 threshold.
	// The starred solo dev gets near-zero Trust (≈0.01) from the Range signal alone (conf 0.72,
	// strength ≈ 0.72). External merged PRs push Trust to 0.93 — and adding more verified
	// evidence can only raise Trust further, never lower it.
	if contrib.Trust <= solo.Trust {
		t.Errorf("external PR contributor Trust %.3f should exceed star-only dev Trust %.3f",
			contrib.Trust, solo.Trust)
	}
	if contrib.Trust < 0.80 {
		t.Errorf("contributor Trust = %.3f, want >= 0.80 (high-confidence external evidence)", contrib.Trust)
	}
}

func TestCompute_Trust_ZeroWithNoSignals(t *testing.T) {
	s := scoring.Compute(zeroInput())
	assertZero(t, "Trust with no signals", s.Trust)
}

func TestCompute_Range_SpecialistConcentration(t *testing.T) {
	spec := scoring.Compute(specialistInput())
	gen := scoring.Compute(generalistInput())

	// Specialist (95% Rust) should have a higher concentration index than generalist.
	if spec.RangeConcentration <= gen.RangeConcentration {
		t.Errorf("specialist RangeConcentration %.3f should be > generalist %.3f",
			spec.RangeConcentration, gen.RangeConcentration)
	}
	// Generalist should approach 0 (perfectly distributed work).
	if gen.RangeConcentration > 0.5 {
		t.Errorf("generalist RangeConcentration %.3f should be <= 0.5", gen.RangeConcentration)
	}
}

func TestCompute_Range_UnvalidatedReposExcluded(t *testing.T) {
	// solodev has one Go repo but 0 stars / 0 external contribs — not validated.
	s := scoring.Compute(soloDevInput())
	if s.Range.Raw > 0 {
		t.Errorf("unvalidated repos must not count for Range: Raw = %v, want 0", s.Range.Raw)
	}
}

func TestCompute_RecencyDecay(t *testing.T) {
	recent := scoring.Compute(recencyInputRecent())
	old := scoring.Compute(recencyInputOld())

	// Same number of active days but one is recent, one is outside or near the window edge.
	if recent.Output.Raw <= old.Output.Raw {
		t.Errorf("recent activity Output.Raw %v should exceed old activity %v",
			recent.Output.Raw, old.Output.Raw)
	}
}

func TestCompute_DurableBuilder_InfluencePersistsDespiteLowOutput(t *testing.T) {
	// Empirical check for per-dimension decay (Decision 3).
	//
	// The durable builder's 5 active days at 300-500 days ago are within the 18-month
	// window but decayed to 17-32% of face value by the 180-day cadence half-life.
	// Output.Raw ≈ 2.1, Output.Percentile ≈ p10.
	//
	// Dependents (5000) and quality-adjusted stars (8000) are queried point-in-time
	// from the GitHub API — they reflect who imports the library today. computeInfluence
	// makes zero decay calls. Influence.Raw ≈ 8384, Influence.Percentile ≈ p99.
	s := scoring.Compute(durableBuilderInput())

	if s.Output.Percentile >= 20 {
		t.Errorf("Output.Percentile = %.1f, want < 20 (5 days at 300-500d should be heavily decayed)",
			s.Output.Percentile)
	}
	if s.Influence.Percentile < 90 {
		t.Errorf("Influence.Percentile = %.1f, want >= 90 (5000 dependents are present-tense, no decay)",
			s.Influence.Percentile)
	}
	// The gap is the key assertion: different half-lives produce different trajectories.
	if s.Influence.Percentile <= s.Output.Percentile {
		t.Errorf("Influence (p%.0f) should far exceed Output (p%.0f): per-dimension decay is not isolating correctly",
			s.Influence.Percentile, s.Output.Percentile)
	}
}

func TestCompute_OutputCap(t *testing.T) {
	// Create an input with an enormous number of active days — raw should be capped.
	days := make([]scoring.ActiveDay, 2000)
	for i := range days {
		days[i] = scoring.ActiveDay{
			RepoOwner:               "dev",
			RepoName:                "repo",
			IsOwnRepo:               true,
			RepoStars:               5000,
			RepoHasExternalContribs: true,
			Date:                    daysAgo(i % 400), // spread across window
		}
	}
	in := scoring.GitHubInput{FetchedAt: now, ActiveDays: days}
	s := scoring.Compute(in)

	if s.Output.Raw > 155 { // 150 cap + small tolerance for floating point
		t.Errorf("Output.Raw = %v exceeds cap of 150", s.Output.Raw)
	}
}

func TestCompute_NormalizeMonotone(t *testing.T) {
	// Percentile rank must be monotonically non-decreasing as raw score increases.
	norm := scoring.DefaultNormalizer()
	prev := -1.0
	for raw := 0.0; raw <= 600; raw += 5 {
		p := norm.Normalize(scoring.DimCollaboration, raw)
		if p < prev {
			t.Errorf("normalization not monotone at raw=%.0f: percentile %.2f < previous %.2f", raw, p, prev)
		}
		prev = p
	}
}

func TestCompute_AllDimensionsHaveSignals_WhenActive(t *testing.T) {
	s := scoring.Compute(openSourceContribInput())

	check := func(name string, ds scoring.DimensionScore) {
		t.Helper()
		if len(ds.Signals) == 0 && ds.Raw > 0 {
			t.Errorf("%s: Raw=%.2f but no Signals recorded", name, ds.Raw)
		}
	}
	check("Output", s.Output)
	check("Craft", s.Craft)
	check("Influence", s.Influence)
	check("Collaboration", s.Collaboration)
	check("Range", s.Range)
}

func TestCompute_Collaboration_SpammerGetsNearZero(t *testing.T) {
	// 100 merged PRs all to 2-star repos with zero review threads.
	// Repos below the minimum reputation threshold earn zero weight — the whole
	// point of the gaming guard is that Hacktoberfest spam doesn't rank a developer.
	spamPRs := make([]scoring.ExternalPR, 100)
	for i := range spamPRs {
		spamPRs[i] = scoring.ExternalPR{
			RepoOwner:         "random-org",
			RepoName:          "repo",
			RepoStars:         2,  // below ~8-star minimum reputation threshold
			ReviewThreadCount: 0,
			MergedAt:          daysAgo(i * 3),
		}
	}
	in := scoring.GitHubInput{FetchedAt: now, ExternalPRs: spamPRs}
	s := scoring.Compute(in)

	// Collaboration must be near zero despite high PR volume.
	if s.Collaboration.Percentile > 10 {
		t.Errorf("spammer Collaboration.Percentile = %.1f, want <= 10 (100 typo-PRs to 2-star repos must not rank)",
			s.Collaboration.Percentile)
	}

	// One substantive PR to a known repo should outrank all 100 spam PRs.
	realPR := scoring.GitHubInput{
		FetchedAt: now,
		ExternalPRs: []scoring.ExternalPR{
			{RepoOwner: "real-org", RepoName: "big-project", RepoStars: 5000,
				ReviewThreadCount: 4, MergedAt: daysAgo(30)},
		},
	}
	sReal := scoring.Compute(realPR)
	if sReal.Collaboration.Percentile <= s.Collaboration.Percentile {
		t.Errorf("one substantive PR (5k-star repo, 4 threads) ranked %.1f but spammer ranked %.1f — gaming guard failed",
			sReal.Collaboration.Percentile, s.Collaboration.Percentile)
	}
}

func TestCompute_Deterministic(t *testing.T) {
	in := openSourceContribInput()
	a := scoring.Compute(in)
	b := scoring.Compute(in)

	if a.Output.Percentile != b.Output.Percentile ||
		a.Trust != b.Trust ||
		a.RangeConcentration != b.RangeConcentration {
		t.Error("Compute is not deterministic: same input produced different output")
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// activeDays generates n active days spread evenly over the last 300 days.
func activeDays(owner, repo string, isOwn bool, stars int, hasExternal bool, deps int, n int) []scoring.ActiveDay {
	if n == 0 {
		return nil
	}
	days := make([]scoring.ActiveDay, n)
	for i := range days {
		days[i] = scoring.ActiveDay{
			RepoOwner:               owner,
			RepoName:                repo,
			IsOwnRepo:               isOwn,
			RepoStars:               stars,
			RepoHasExternalContribs: hasExternal,
			RepoDependentCount:      deps,
			Date:                    daysAgo(i * 300 / max(n, 1)),
		}
	}
	return days
}

// activeDaysSpread generates n active days between startDay and endDay days ago.
func activeDaysSpread(owner, repo string, isOwn bool, stars int, hasExternal bool, deps int, startDay, endDay, n int) []scoring.ActiveDay {
	if n == 0 {
		return nil
	}
	days := make([]scoring.ActiveDay, n)
	span := endDay - startDay
	for i := range days {
		offset := startDay + (i*span)/max(n, 1)
		days[i] = scoring.ActiveDay{
			RepoOwner:               owner,
			RepoName:                repo,
			IsOwnRepo:               isOwn,
			RepoStars:               stars,
			RepoHasExternalContribs: hasExternal,
			RepoDependentCount:      deps,
			Date:                    daysAgo(offset),
		}
	}
	return days
}

func assertZero(t *testing.T, label string, v float64) {
	t.Helper()
	if v != 0 {
		t.Errorf("%s = %v, want 0", label, v)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
