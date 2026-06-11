package scoring

import (
	"fmt"
	"math"
)

// computeInfluence scores Influence / Reach.
//
// Gold signal: package dependents — others literally import your code.
// Strong: stars WITH forks + external contributors + issues (hard to fake in combination).
// Weak: stars alone (partially gameable), forks alone.
// Star quality is assessed per-repo; a spike of stars with no forks/issues/contributors
// gets a low quality multiplier.
func computeInfluence(in GitHubInput) (raw float64, sigs []Signal) {
	var (
		starPts float64
		depPts  float64
		forkPts float64
	)

	const perRepoCap = 400.0

	for _, r := range in.OwnedRepos {
		// Quality multiplier: stars supported by other evidence are more real.
		// 0.40 base + bonuses for forks, external contributors, open issues.
		q := 0.40
		if r.Forks > 0 {
			q += 0.20
		}
		if r.ExternalContributors > 0 {
			q += 0.30
		}
		if r.OpenIssues > 0 {
			q += 0.10
		}
		// q ∈ [0.40, 1.00]

		repoContrib := 0.0

		// Stars (quality-adjusted).
		if r.Stars > 0 {
			contribution := float64(r.Stars) * q
			starPts += contribution
			repoContrib += contribution
		}

		// Dependents: gold signal, log-scaled so viral packages don't blow up the raw.
		// log2 scale: 1 dep → ~1, 10 deps → ~3.3, 100 deps → ~6.6, 10k deps → ~13.3, 1M deps → ~20
		if r.DependentCount > 0 {
			pts := 15.0 * math.Log2(float64(r.DependentCount)+1)
			depPts += pts
			repoContrib += pts
		}

		// Forks (secondary; shows others forked to build on or study).
		if r.Forks > 0 {
			pts := 1.5 * float64(r.Forks)
			forkPts += pts
			repoContrib += pts
		}

		_ = math.Min(repoContrib, perRepoCap) // per-repo cap tracked but raw totals used below
	}

	// Per-source caps on the totals.
	const forkCap = 200.0
	forkPts = math.Min(forkPts, forkCap)

	type sigSpec struct {
		label      string
		pts        float64
		confidence float64
	}
	specs := []sigSpec{
		{"quality-adjusted stars on owned repos", starPts, qualifiedStarConfidence(in)},
		{"package dependents (npm / PyPI / crates.io)", depPts, 0.92},
		{fmt.Sprintf("forks of owned repos (capped at %.0f)", forkCap), forkPts, 0.75},
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

// qualifiedStarConfidence computes an aggregate confidence for the star signal.
// Repos with forks + contributors → higher confidence; stars-only → lower.
func qualifiedStarConfidence(in GitHubInput) float64 {
	totalStars, qualifiedStars := 0, 0
	for _, r := range in.OwnedRepos {
		totalStars += r.Stars
		if r.Forks > 0 || r.ExternalContributors > 0 {
			qualifiedStars += r.Stars
		}
	}
	if totalStars == 0 {
		return 0.60
	}
	// Weighted blend: 0.55 (stars-only) to 0.82 (all stars have supporting evidence).
	qualRatio := float64(qualifiedStars) / float64(totalStars)
	return 0.55 + 0.27*qualRatio
}
