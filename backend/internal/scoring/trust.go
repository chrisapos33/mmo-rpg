package scoring

// computeTrust derives the Trust meta-score from high-confidence dimension signals.
//
// Model: accumulate "verified strength" = Σ(pts × confidence) for every signal with
// confidence ≥ 0.70, then pass it through a saturating hyperbolic function:
//
//	Trust = strength / (strength + halfSaturation)
//
// This is monotonic by construction: d(Trust)/d(strength) = k/(strength+k)² > 0.
// Any additional verified signal increases strength, which always increases Trust.
// The weighted-average formula fails this because low-conf signals dilute the mean.
//
// Half-saturation k ≈ 70 is calibrated so that:
//   - strength ≈ 70 (a handful of solid external PRs) → Trust ≈ 0.50
//   - strength ≈ 280 (crafted dev: reviewed PRs, validated stars, deps) → Trust ≈ 0.80
//   - strength ≈ 975 (top OSS contributor: 6 PRs to mega-repos + reviews + deps) → Trust ≈ 0.93
//
// Confidence threshold at 0.70: signals below it (tests, CI, longevity, stars-only
// activity) contribute to dimension percentile ranks but never to Trust.
func computeTrust(dimSigs ...[]Signal) float64 {
	const (
		minConfidence  = 0.70
		halfSaturation = 70.0
	)

	var strength float64
	for _, sigs := range dimSigs {
		for _, s := range sigs {
			if s.Points > 0 && s.Confidence >= minConfidence {
				strength += s.Points * s.Confidence
			}
		}
	}

	return strength / (strength + halfSaturation)
}
