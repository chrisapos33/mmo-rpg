package scoring

import "time"

// Compute derives dimension scores and Trust from raw GitHub input using the
// default (seeded reference) normalizer. Same input always produces the same output.
func Compute(in GitHubInput) Scores {
	return ComputeWith(in, DefaultNormalizer())
}

// ComputeWith is the same as Compute but accepts a custom normalizer.
// Use this to swap in a cohort-percentile normalizer once the user base is large enough.
func ComputeWith(in GitHubInput, norm Normalizer) Scores {
	now := in.FetchedAt
	if now.IsZero() {
		now = time.Now()
	}

	outputRaw, outputSigs := computeCadence(in, now)
	craftRaw, craftSigs := computeCraft(in, now)
	influenceRaw, influenceSigs := computeInfluence(in)
	collabRaw, collabSigs := computeCollab(in, now)
	rangeRaw, rangeSigs, rangeConc := computeBreadth(in)

	trust := computeTrust(outputSigs, craftSigs, influenceSigs, collabSigs, rangeSigs)

	return Scores{
		Output: DimensionScore{
			Raw:        outputRaw,
			Percentile: norm.Normalize(DimOutput, outputRaw),
			Signals:    outputSigs,
		},
		Craft: DimensionScore{
			Raw:        craftRaw,
			Percentile: norm.Normalize(DimCraft, craftRaw),
			Signals:    craftSigs,
		},
		Influence: DimensionScore{
			Raw:        influenceRaw,
			Percentile: norm.Normalize(DimInfluence, influenceRaw),
			Signals:    influenceSigs,
		},
		Collaboration: DimensionScore{
			Raw:        collabRaw,
			Percentile: norm.Normalize(DimCollaboration, collabRaw),
			Signals:    collabSigs,
		},
		Range: DimensionScore{
			Raw:        rangeRaw,
			Percentile: norm.Normalize(DimRange, rangeRaw),
			Signals:    rangeSigs,
		},
		RangeConcentration: rangeConc,
		Trust:              trust,
		ComputedAt:         now,
	}
}
