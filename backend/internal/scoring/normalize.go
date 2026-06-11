package scoring

import "math"

// DimID identifies a scoring dimension for normalization lookup.
type DimID int

const (
	DimOutput        DimID = iota
	DimCraft               // Craft / Quality
	DimInfluence           // Influence / Reach
	DimCollaboration       // Collaboration
	DimRange               // Range (language depth)
)

// Normalizer converts a raw dimension score to a 0–100 percentile rank.
// Swap implementations: ReferenceNormalizer (cold-start) or a future
// CohortNormalizer once the user base is large enough for real percentiles.
type Normalizer interface {
	Normalize(dim DimID, raw float64) float64
}

// DefaultNormalizer returns the seeded reference-distribution normalizer.
func DefaultNormalizer() Normalizer { return refNorm{} }

// refNorm uses the hand-seeded reference distribution for cold-start normalization.
type refNorm struct{}

func (refNorm) Normalize(dim DimID, raw float64) float64 {
	pts, ok := refBreakpoints[dim]
	if !ok || len(pts) == 0 {
		return 0
	}
	return math.Min(interpolate(pts, raw), 100)
}

// refPoint is one (raw score, percentile) sample in the reference distribution.
type refPoint struct {
	raw        float64
	percentile float64
}

// interpolate does piecewise-linear interpolation between adjacent breakpoints.
func interpolate(pts []refPoint, raw float64) float64 {
	if raw <= pts[0].raw {
		return pts[0].percentile
	}
	last := pts[len(pts)-1]
	if raw >= last.raw {
		return last.percentile
	}
	for i := 1; i < len(pts); i++ {
		if raw <= pts[i].raw {
			lo, hi := pts[i-1], pts[i]
			t := (raw - lo.raw) / (hi.raw - lo.raw)
			return lo.percentile + t*(hi.percentile-lo.percentile)
		}
	}
	return last.percentile
}
