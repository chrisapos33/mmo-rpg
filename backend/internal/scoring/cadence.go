package scoring

import (
	"fmt"
	"math"
	"time"
)

// computeCadence scores Output / Cadence.
//
// Counts distinct active days on externally-validated repos with recency decay.
// Unvalidated repos (zero stars, no external contributors, no dependents) are
// excluded entirely — activity in a vacuum signals nothing. This is table-stakes
// (answers "are they alive?"), not a ranker, so the raw score is capped.
func computeCadence(in GitHubInput, now time.Time) (raw float64, sigs []Signal) {
	// Three validation tiers, each with its own quality multiplier and confidence.
	var (
		depWeight, comWeight, starWeight float64
	)

	// Deduplicate: one contribution per (repo, calendar-day).
	type key struct{ owner, name, date string }
	seen := make(map[key]bool, len(in.ActiveDays))

	for _, d := range in.ActiveDays {
		k := key{d.RepoOwner, d.RepoName, d.Date.Format("2006-01-02")}
		if seen[k] {
			continue
		}
		seen[k] = true

		w := decayWeight(d.Date, now)
		if w == 0 {
			continue
		}

		// Popularity boost for highly-starred repos (log scale, additive).
		starBoost := 1.0
		if d.RepoStars >= 100 {
			starBoost = 1.0 + math.Log10(float64(d.RepoStars))/10.0
		}

		switch {
		case d.RepoDependentCount > 0:
			depWeight += math.Min(w*1.30*starBoost, 1.5)
		case d.RepoHasExternalContribs:
			comWeight += math.Min(w*1.10*starBoost, 1.5)
		case d.RepoStars > 0:
			starWeight += math.Min(w*0.80*starBoost, 1.5)
		default:
			// Unvalidated: no external evidence anyone cared about this repo.
		}
	}

	const rawCap = 150.0

	type sigSpec struct {
		label      string
		weight     float64
		confidence float64
	}
	specs := []sigSpec{
		{"active days on dependency-validated repos", depWeight, 0.88},
		{"active days on community-validated repos", comWeight, 0.78},
		{"active days on star-validated repos", starWeight, 0.62},
	}

	total := 0.0
	for _, sp := range specs {
		if sp.weight <= 0 {
			continue
		}
		sigs = append(sigs, Signal{
			Description: fmt.Sprintf("%.1f decayed %s", sp.weight, sp.label),
			Points:      sp.weight,
			Confidence:  sp.confidence,
		})
		total += sp.weight
	}

	raw = math.Min(total, rawCap)

	// Scale signal points proportionally if the global cap was hit.
	if total > rawCap && len(sigs) > 0 {
		scale := rawCap / total
		for i := range sigs {
			sigs[i].Points *= scale
		}
	}

	return raw, sigs
}
