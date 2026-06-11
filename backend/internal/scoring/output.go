package scoring

import "time"

// Scores is the computed result for one developer's GitHub profile.
// All Percentile fields are 0–100 ranks within the reference cohort.
// Trust is a separate 0.0–1.0 meta-score, never ranked.
type Scores struct {
	Output        DimensionScore // Output / Cadence
	Craft         DimensionScore // Craft / Quality
	Influence     DimensionScore // Influence / Reach
	Collaboration DimensionScore // Collaboration
	Range         DimensionScore // Range (language depth distribution)

	// RangeConcentration is 0.0 (pure generalist) → 1.0 (pure specialist).
	// This is the Herfindahl–Hirschman index of the validated language distribution.
	// Used by the AI layer to pick class flavor; not exposed as a ranked score.
	RangeConcentration float64

	// Trust is the weighted-average confidence of all evidence in this build.
	// 0 = entirely self-reported / unverifiable.
	// 1 = entirely third-party attributed (someone else merged your PR, your package has dependents).
	// Increases as the user connects real verified signals.
	Trust float64

	ComputedAt time.Time
}

// DimensionScore holds both the 0–100 percentile rank and the underlying evidence.
type DimensionScore struct {
	Percentile float64  // 0–100 rank within the reference cohort
	Raw        float64  // pre-normalization value; inspect for debugging or display
	Signals    []Signal // individual evidence items that fed this score
}

// Signal is one piece of evidence contributing to a dimension score.
type Signal struct {
	Description string
	Points      float64 // raw contribution to this dimension (before normalization)
	Confidence  float64 // 0.0–1.0; third-party-attributed evidence scores higher
}
