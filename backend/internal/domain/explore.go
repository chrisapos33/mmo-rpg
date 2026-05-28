package domain

import (
	"time"

	"github.com/lib/pq"
)

// ExploreEntry is the compact profile card returned by the explore endpoint.
// It joins profiles + user_signal_scores + github_connections in a single query.
type ExploreEntry struct {
	UserID         int64          `db:"user_id"         json:"user_id"`
	Class          *string        `db:"class"           json:"class"`
	Subclass       *string        `db:"subclass"        json:"subclass"`
	Headline       *string        `db:"headline"        json:"headline"`
	TotalSignal    int            `db:"total_signal"    json:"total_signal"`
	GitHubUsername *string        `db:"github_username" json:"github_username"`
	TopLanguages   pq.StringArray `db:"top_languages"   json:"top_languages"`
	UpdatedAt      time.Time      `db:"updated_at"      json:"updated_at"`
}

// AllClasses is the canonical ordered list for UI filter pills.
var AllClasses = []string{
	"The Architect",
	"The Artisan",
	"The Pathfinder",
	"The Sage",
	"The Operator",
	"The Sentinel",
	"The Artificer",
}
