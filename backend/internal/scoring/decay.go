package scoring

import (
	"math"
	"time"
)

const (
	// Output/Cadence: fast decay — "are they active NOW?" 6-month half-life.
	decayHalfLifeOutput = 180.0

	// Collaboration: medium decay — reputation from external PRs persists ~12 months.
	decayHalfLifeCollab = 365.0

	// Craft: slow decay — review depth is evidence of craft quality, not current activity.
	decayHalfLifeCraft = 730.0

	// windowDays: activity older than ~18 months contributes nothing.
	windowDays = 548.0
)

// decayWeightWith returns a [0, 1] multiplier for an event at time t with a custom half-life.
func decayWeightWith(halfLife float64, t, now time.Time) float64 {
	daysAgo := now.Sub(t).Hours() / 24.0
	if daysAgo < 0 {
		daysAgo = 0
	}
	if daysAgo > windowDays {
		return 0
	}
	return math.Pow(2, -daysAgo/halfLife)
}

// decayWeight returns a [0, 1] multiplier using the Output/Cadence half-life (180 days).
// Use decayWeightWith for Collaboration (365) or Craft (730).
func decayWeight(t, now time.Time) float64 {
	return decayWeightWith(decayHalfLifeOutput, t, now)
}

// inWindow reports whether event time t falls within the active scoring window.
func inWindow(t, now time.Time) bool {
	return now.Sub(t).Hours()/24.0 <= windowDays
}
