package scoring

import (
	"fmt"
	"math"
	"time"
)

// computeCollab scores Collaboration.
//
// Gold signal: merged PRs to repos the user does NOT own. A maintainer approved
// the work — the user cannot unilaterally merge into someone else's repo, so this
// is expensive to fake. Weighted by repo popularity and PR substance (review threads).
//
// Gaming guard: typo-PRs and Hacktoberfest spam land on tiny repos with zero review
// — the repo weight (log scale on stars) and substance multiplier (review threads)
// together ensure one substantive PR to a known repo >> 50 trivial PRs.
//
// Secondary: reviews given to others' PRs (caps applied to prevent mass-review gaming).
func computeCollab(in GitHubInput, now time.Time) (raw float64, sigs []Signal) {
	const (
		perPRCap    = 60.0 // single PR can't dominate
		reviewCap   = 80.0 // total review contribution cap
		reviewPtEach = 8.0
	)

	var prPts, reviewPts float64

	for _, pr := range in.ExternalPRs {
		if !inWindow(pr.MergedAt, now) {
			continue
		}
		decay := decayWeightWith(decayHalfLifeCollab, pr.MergedAt, now)

		// Repo weight: requires a minimum reputation before a PR earns any credit.
		// Threshold ≈ log10(9) = 0.95, so repos with fewer than ~8 stars earn zero.
		// Above the threshold it grows with log scale — prevents viral repos from
		// creating an insurmountable gap while still rewarding popular-project merges.
		// At 50 stars ≈ 0.56, at 1k stars ≈ 1.4, at 100k stars ≈ 2.7.
		repoWeight := math.Max(0, (math.Log10(float64(pr.RepoStars+1))-0.95)/1.5)

		// Substance multiplier: 0.5 base (a merge always counts for something), grows
		// up to 2.0 at 5+ review threads. A reviewed PR to a real repo is worth much
		// more than a drive-by one-liner with zero discussion.
		substance := 0.5 + math.Min(float64(pr.ReviewThreadCount), 5.0)*0.3

		pts := math.Min(25.0*repoWeight*substance*decay, perPRCap)
		prPts += pts
	}

	for _, rv := range in.ReviewsGiven {
		if !inWindow(rv.CreatedAt, now) {
			continue
		}
		decay := decayWeightWith(decayHalfLifeCollab, rv.CreatedAt, now)
		reviewPts += reviewPtEach * decay
	}
	reviewPts = math.Min(reviewPts, reviewCap)

	// Cross-org breadth: distinct organizations the user contributed external PRs to.
	orgSet := make(map[string]bool)
	for _, pr := range in.ExternalPRs {
		if inWindow(pr.MergedAt, now) && pr.RepoOwner != "" {
			orgSet[pr.RepoOwner] = true
		}
	}
	breadthBonus := 0.0
	if n := len(orgSet); n >= 3 {
		breadthBonus = math.Min(float64(n-2)*5.0, 30.0) // +5 per org above 2, capped
	}

	type sigSpec struct {
		label      string
		pts        float64
		confidence float64
	}
	specs := []sigSpec{
		{
			fmt.Sprintf("%d merged PRs to repos you don't own (weighted by popularity + review depth)",
				countWindowedPRs(in, now)),
			prPts,
			0.92,
		},
		{
			fmt.Sprintf("code reviews given to other contributors (capped at %.0f pts)", reviewCap),
			reviewPts,
			0.72,
		},
		{
			fmt.Sprintf("cross-org breadth bonus (%d distinct orgs contributed to)", len(orgSet)),
			breadthBonus,
			0.85,
		},
	}

	total := 0.0
	for _, sp := range specs {
		if sp.pts <= 0 {
			continue
		}
		sigs = append(sigs, Signal{
			Description: sp.label,
			Points:      sp.pts,
			Confidence:  sp.confidence,
		})
		total += sp.pts
	}

	raw = total
	return raw, sigs
}

func countWindowedPRs(in GitHubInput, now time.Time) int {
	n := 0
	for _, pr := range in.ExternalPRs {
		if inWindow(pr.MergedAt, now) {
			n++
		}
	}
	return n
}
