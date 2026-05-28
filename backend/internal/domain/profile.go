package domain

import (
	"time"

	"github.com/lib/pq"
)

type Profile struct {
	ID             int64          `db:"id"              json:"id"`
	UserID         int64          `db:"user_id"         json:"user_id"`
	Username       *string        `db:"username"        json:"username"`
	DisplayName    *string        `db:"display_name"    json:"display_name"`
	Class          *string        `db:"class"           json:"class"`
	Subclass       *string        `db:"subclass"        json:"subclass"`
	Headline       *string        `db:"headline"        json:"headline"`
	Summary        *string        `db:"summary"         json:"summary"`
	AvatarURL      *string        `db:"avatar_url"      json:"avatar_url"`
	SignalScore    int            `db:"signal_score"    json:"signal_score"`
	XP             int            `db:"xp"              json:"xp"`
	IsPublished    bool           `db:"is_published"    json:"is_published"`
	OnboardingStep string         `db:"onboarding_step" json:"onboarding_step"`
	Strengths      pq.StringArray `db:"strengths"       json:"strengths"`
	GrowthPaths    pq.StringArray `db:"growth_paths"    json:"growth_paths"`
	CreatedAt      time.Time      `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"      json:"updated_at"`
}

// BuildData is the structured output from AI build generation.
type BuildData struct {
	Class       string   `json:"class"`
	Subclass    string   `json:"subclass"`
	Headline    string   `json:"headline"`
	Summary     string   `json:"summary"`
	Strengths   []string `json:"strengths"`
	GrowthPaths []string `json:"growth_paths"`
}
