package scoring

// refBreakpoints defines the seeded percentile distribution for each dimension.
//
// TARGET POPULATION — critical for anyone recalibrating this table:
// The reference cohort is ACTIVE DEVELOPERS who can demonstrate their level —
// people who ship code publicly, contribute to OSS, or maintain real projects.
// It is NOT all GitHub accounts (most of which have minimal public activity).
// The median breakpoint (p50) should reflect a developer who is meaningfully
// engaged: a few maintained repos, occasional external contributions, some
// community presence. A solo dev with one hobby project should land below p50;
// a developer with validated OSS work and collaborator history should land well
// above it. Do NOT calibrate against the full GitHub user distribution — that
// would compress everyone who matters into the top 10%.
//
// Replace with real cohort data (see CohortNormalizer) once the user base
// reaches ~500 profiles from which a real distribution can be derived.
//
// Reading the table: a raw Collaboration score of 55 puts a developer at the 75th
// percentile — only 25% of the reference cohort scores higher.
var refBreakpoints = map[DimID][]refPoint{
	// Output / Cadence: decayed active days on externally-validated repos.
	// Most devs have limited activity on repos others have actually validated.
	DimOutput: {
		{0, 0},
		{0.5, 5},
		{2, 10},
		{8, 25},
		{20, 50},
		{45, 75},
		{80, 90},
		{110, 95},
		{180, 99},
		{300, 100},
	},

	// Craft / Quality: test presence + CI + review depth + longevity.
	// The review depth signal requires others to engage — medium rarity.
	DimCraft: {
		{0, 0},
		{1, 10},
		{5, 25},
		{16, 50},
		{38, 75},
		{65, 90},
		{85, 95},
		{120, 99},
		{200, 100},
	},

	// Influence / Reach: quality-adjusted stars + dependents + forks.
	// Highly skewed: most devs have <50 total stars; viral OSS is rare.
	DimInfluence: {
		{0, 0},
		{0.5, 10},
		{2, 25},
		{8, 50},
		{45, 75},
		{200, 90},
		{500, 95},
		{2000, 99},
		{20000, 100},
	},

	// Collaboration: external merged PRs (weighted by reputation + review depth) + reviews given.
	// Most developers never merge a PR to a repo they don't own. The gaming guard
	// (minimum star threshold) means Hacktoberfest / typo-PR volume earns near zero.
	// A few substantive PRs to known repos (50+ stars, 2+ threads) push into the 70s.
	DimCollaboration: {
		{0, 0},
		{1, 30},
		{8, 50},
		{40, 75},
		{130, 90},
		{280, 95},
		{600, 99},
		{5000, 100},
	},

	// Range: total validated language depth (quality-factor × language-share, summed across
	// all validated repos). Each validated repo contributes at most quality (≈1–2) to the
	// raw total. Reference thresholds reflect: p25 ≈ 1 validated repo (0.5–1 raw),
	// p50 ≈ 3 repos (2–3 raw), p75 ≈ 5 repos (6–8 raw), p90 ≈ 10 repos (15–20 raw).
	DimRange: {
		{0, 0},
		{0.2, 10},
		{0.7, 25},
		{2.5, 50},
		{7.0, 75},
		{18.0, 90},
		{30.0, 95},
		{60.0, 99},
		{200.0, 100},
	},
}
