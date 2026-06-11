package scoring

import (
	"fmt"
	"math"
	"time"
)

// computeCraft scores Craft / Quality.
//
// Proxies: test file presence, CI config, repo longevity, and depth of review
// the user's PRs receive (hard to fake — involves other people).
//
// Self-controlled signals (tests, CI, longevity) are gated by repoValidationFactor:
// a solo repo with no stars, forks, or external contributors scores near zero.
// The same signals on a validated repo (480★, external contributors) earn full weight.
// This prevents a developer from ranking themselves by writing CI configs alone.
func computeCraft(in GitHubInput, now time.Time) (raw float64, sigs []Signal) {
	var testPts, ciPts, longevityPts, reviewPts float64

	for _, r := range in.OwnedRepos {
		vf := repoValidationFactor(r)

		repoCap := 80.0
		repoTotal := 0.0

		if r.HasTests {
			pts := math.Min(20.0*vf, repoCap-repoTotal)
			repoTotal += pts
			testPts += pts
		}
		if r.HasCI {
			pts := math.Min(15.0*vf, repoCap-repoTotal)
			repoTotal += pts
			ciPts += pts
		}

		// Longevity: repo maintained over multiple years with recent activity.
		// Also gated by vf — an old abandoned solo repo is not a craft signal.
		recentlyActive := now.Sub(r.UpdatedAt).Hours() < 24*182
		if recentlyActive && !r.CreatedAt.IsZero() {
			yearsOld := now.Sub(r.CreatedAt).Hours() / (24.0 * 365.0)
			if yearsOld >= 1.0 {
				pts := math.Min(10.0*math.Min(yearsOld, 3.0)*vf, repoCap-repoTotal)
				repoTotal += pts
				longevityPts += pts
			}
		}
		_ = repoTotal
	}

	// Review depth: decay-weighted average review threads across windowed external PRs.
	// Already fully third-party gated (requires other people to engage), so no vf needed.
	{
		var totalWeight, weightedThreads float64
		for _, pr := range in.ExternalPRs {
			if !inWindow(pr.MergedAt, now) {
				continue
			}
			d := decayWeightWith(decayHalfLifeCraft, pr.MergedAt, now)
			totalWeight += d
			weightedThreads += float64(pr.ReviewThreadCount) * d
		}
		if totalWeight > 0 {
			avgThreads := weightedThreads / totalWeight
			reviewPts = 15.0 * math.Min(avgThreads, 5.0)
		}
	}

	const rawCap = 250.0

	type sigSpec struct {
		label      string
		pts        float64
		confidence float64
	}
	specs := []sigSpec{
		{"test files present in validated owned repos", testPts, 0.58},
		{"CI config present in validated owned repos", ciPts, 0.55},
		{"validated repos actively maintained over multiple years", longevityPts, 0.52},
		{fmt.Sprintf("avg %.1f review threads per authored external PR", reviewDepthAvg(in, now)), reviewPts, 0.88},
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

	raw = math.Min(total, rawCap)
	return raw, sigs
}

// repoValidationFactor gates craft signals by third-party repo engagement.
// Returns 0.0 for a fully solo unvalidated repo (no stars, no forks, no external
// contributors) and up to 2.0 for highly validated repos.
//
// Gate opens via the strongest available third-party signal (max, not sum):
//   - byStars: log-scale, reaches 1.0 at ~512 stars
//   - byContribs: linear, reaches 1.0 at 5 external contributors
//   - byForks: log-scale, reaches 1.0 at ~128 forks
//
// Factor curve: gate × (1 + gate) — smooth 0→0, 0.5→0.75, 1.0→2.0.
// Amplifies exceptional repos above 1× without letting unvalidated repos
// escape the zero floor.
func repoValidationFactor(r OwnedRepo) float64 {
	byStars := math.Min(math.Log2(float64(r.Stars+1))/9.0, 1.0)
	byContribs := math.Min(float64(r.ExternalContributors)/5.0, 1.0)
	byForks := math.Min(math.Log2(float64(r.Forks+1))/7.0, 1.0)
	gate := math.Max(byStars, math.Max(byContribs, byForks))
	return gate * (1.0 + gate)
}

func reviewDepthAvg(in GitHubInput, now time.Time) float64 {
	var totalWeight, weightedThreads float64
	for _, pr := range in.ExternalPRs {
		if !inWindow(pr.MergedAt, now) {
			continue
		}
		d := decayWeightWith(decayHalfLifeCraft, pr.MergedAt, now)
		totalWeight += d
		weightedThreads += float64(pr.ReviewThreadCount) * d
	}
	if totalWeight == 0 {
		return 0
	}
	return weightedThreads / totalWeight
}
