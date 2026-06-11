package scoring_test

import (
	"fmt"
	"testing"

	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

func TestSanityPrint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	printScores := func(label string, s scoring.Scores) {
		fmt.Printf("\n═══ %s ═══\n", label)
		fmt.Printf("  Output:        raw=%5.1f  p=%4.0f\n", s.Output.Raw, s.Output.Percentile)
		fmt.Printf("  Craft:         raw=%5.1f  p=%4.0f\n", s.Craft.Raw, s.Craft.Percentile)
		fmt.Printf("  Influence:     raw=%5.1f  p=%4.0f\n", s.Influence.Raw, s.Influence.Percentile)
		fmt.Printf("  Collaboration: raw=%5.1f  p=%4.0f\n", s.Collaboration.Raw, s.Collaboration.Percentile)
		fmt.Printf("  Range:         raw=%5.1f  p=%4.0f  conc=%.2f\n", s.Range.Raw, s.Range.Percentile, s.RangeConcentration)
		fmt.Printf("  Trust:         %.2f\n", s.Trust)
		for _, sig := range s.Output.Signals        { fmt.Printf("    [output]      %s  pts=%.1f conf=%.2f\n", sig.Description, sig.Points, sig.Confidence) }
		for _, sig := range s.Craft.Signals         { fmt.Printf("    [craft]       %s  pts=%.1f conf=%.2f\n", sig.Description, sig.Points, sig.Confidence) }
		for _, sig := range s.Influence.Signals     { fmt.Printf("    [influence]   %s  pts=%.1f conf=%.2f\n", sig.Description, sig.Points, sig.Confidence) }
		for _, sig := range s.Collaboration.Signals { fmt.Printf("    [collab]      %s  pts=%.1f conf=%.2f\n", sig.Description, sig.Points, sig.Confidence) }
		for _, sig := range s.Range.Signals         { fmt.Printf("    [range]       %s  pts=%.1f conf=%.2f\n", sig.Description, sig.Points, sig.Confidence) }
	}

	printScores("Zero Input", scoring.Compute(zeroInput()))
	printScores("Solo Dev (0 stars, nothing validated)", scoring.Compute(soloDevInput()))
	printScores("Starred Solo Dev (50 stars, no external)", scoring.Compute(starredSoloDevInput()))
	printScores("Crafted Dev (tests+CI+reviewed PRs)", scoring.Compute(craftedDevInput()))
	printScores("Open Source Contributor (top OSS)", scoring.Compute(openSourceContribInput()))
	printScores("Durable Builder (popular lib, minimal recent commits)", scoring.Compute(durableBuilderInput()))
	printScores("Specialist (95% Rust)", scoring.Compute(specialistInput()))
	printScores("Generalist (5 equal languages)", scoring.Compute(generalistInput()))

	// Spammer: 100 PRs to 2-star repos, zero review threads
	spamPRs := make([]scoring.ExternalPR, 100)
	for i := range spamPRs {
		spamPRs[i] = scoring.ExternalPR{
			RepoOwner: "random", RepoName: fmt.Sprintf("repo-%d", i),
			RepoStars: 2, ReviewThreadCount: 0, MergedAt: daysAgo(i * 3),
		}
	}
	printScores("Hacktoberfest Spammer (100 PRs, 2-star repos, 0 threads)",
		scoring.Compute(scoring.GitHubInput{FetchedAt: now, ExternalPRs: spamPRs}))
}
