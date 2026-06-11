package scoring

import (
	"fmt"
	"math"
	"sort"
)

// computeBreadth scores Range — the language depth dimension.
//
// NOT a count of languages. Measures depth of externally-validated work per language:
// repos with zero stars, zero external contributors, and zero dependents don't count
// — depth in an unvalidated silo doesn't demonstrate range.
//
// Returns:
//   - raw: total validated language depth (feeds percentile normalization)
//   - sigs: evidence signals with description/points/confidence
//   - concentration: HHI index [0, 1] — 1 = pure specialist, 0 = pure generalist.
//     Used by the AI layer for class flavor; not ranked.
func computeBreadth(in GitHubInput) (raw float64, sigs []Signal, concentration float64) {
	// langDepth[lang] accumulates quality-weighted depth across validated repos.
	langDepth := make(map[string]float64)
	validatedRepos := 0

	for _, r := range in.OwnedRepos {
		validated := r.Stars > 0 || r.ExternalContributors > 0 || r.DependentCount > 0
		if !validated {
			continue
		}
		validatedRepos++

		totalBytes := int64(0)
		for _, b := range r.Languages {
			totalBytes += b
		}
		if totalBytes == 0 {
			continue
		}

		// Quality factor: more-starred/depended-upon repos carry more weight.
		// Ranges from 0.5 (barely validated) to 2.0 (very popular).
		q := 0.5 + math.Min(float64(r.Stars)*0.01+float64(r.DependentCount)*0.1, 1.5)

		for lang, bytes := range r.Languages {
			share := float64(bytes) / float64(totalBytes)
			langDepth[lang] += share * q
		}
	}

	// Compute total depth and HHI (Herfindahl–Hirschman Index).
	totalDepth := 0.0
	for _, d := range langDepth {
		totalDepth += d
	}

	if totalDepth == 0 {
		return 0, nil, 0
	}

	hhi := 0.0
	for _, d := range langDepth {
		share := d / totalDepth
		hhi += share * share
	}

	// Build sorted top-languages list for the signal description.
	type langScore struct {
		lang  string
		depth float64
	}
	langs := make([]langScore, 0, len(langDepth))
	for l, d := range langDepth {
		langs = append(langs, langScore{l, d})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].depth > langs[j].depth })

	topN := min(3, len(langs))
	topLabels := make([]string, topN)
	for i := range topLabels {
		topLabels[i] = langs[i].lang
	}

	sigs = []Signal{
		{
			Description: fmt.Sprintf("validated language depth across %d repos — top: %v",
				validatedRepos, topLabels),
			Points:     totalDepth,
			Confidence: 0.72,
		},
	}

	return totalDepth, sigs, hhi
}
