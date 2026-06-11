package domain

import (
	"encoding/json"
	"time"
)

// Source types for evidence items.
const (
	SourceGitHub    = "github"
	SourceBlog      = "blog"
	SourcePortfolio = "portfolio"
	SourceCommunity = "community"
	SourceLinkedIn  = "linkedin"
	SourceManual    = "manual"
	SourceOther     = "other"
)

// Verification statuses (ordered by confidence level).
const (
	VerifUnverified       = "unverified"
	VerifURLVerified      = "url_verified"
	VerifPlatformVerified = "platform_verified"
	VerifPeerVerified     = "peer_verified"
	VerifAdminVerified    = "admin_verified"
)

// Signal dimensions used in signal_events.dimension (evidence panel taxonomy).
// These are NOT the scoring engine dimensions — they power the evidence audit trail.
const (
	DimBuilder      = "builder"
	DimThinker      = "thinker"
	DimExecutor     = "executor"
	DimCollaborator = "collaborator"
	DimSpecialist   = "specialist"
)

// EvidenceItem represents a verifiable artifact from an external source.
type EvidenceItem struct {
	ID                     int64            `db:"id"                      json:"id"`
	UserID                 int64            `db:"user_id"                 json:"user_id"`
	SourceType             string           `db:"source_type"             json:"source_type"`
	SourceKey              string           `db:"source_key"              json:"source_key"`
	ArtifactURL            *string          `db:"artifact_url"            json:"artifact_url"`
	Title                  string           `db:"title"                   json:"title"`
	Description            *string          `db:"description"             json:"description"`
	MetadataJSON           *json.RawMessage `db:"metadata_json"           json:"metadata"`
	VerificationStatus     string           `db:"verification_status"     json:"verification_status"`
	VerificationConfidence float64          `db:"verification_confidence" json:"verification_confidence"`
	CreatedAt              time.Time        `db:"created_at"              json:"created_at"`
	UpdatedAt              time.Time        `db:"updated_at"              json:"updated_at"`
}

// SignalEvent records points awarded to a dimension from a specific evidence item.
type SignalEvent struct {
	ID                   int64     `db:"id"                    json:"id"`
	UserID               int64     `db:"user_id"               json:"user_id"`
	EvidenceItemID       *int64    `db:"evidence_item_id"      json:"evidence_item_id"`
	Dimension            string    `db:"dimension"             json:"dimension"`
	BasePoints           int       `db:"base_points"           json:"base_points"`
	WeightMultiplier     float64   `db:"weight_multiplier"     json:"weight_multiplier"`
	ConfidenceMultiplier float64   `db:"confidence_multiplier" json:"confidence_multiplier"`
	FinalPoints          int       `db:"final_points"          json:"final_points"`
	Explanation          *string   `db:"explanation"           json:"explanation"`
	CreatedAt            time.Time `db:"created_at"            json:"created_at"`
}

// UserSignalScore is the scoring-engine output persisted per user.
// All score columns are written by the background scoring job via scoring.Compute.
// The ScoringStatus* columns track the job lifecycle so the frontend can poll,
// mirroring the cv_uploads.status pattern.
type UserSignalScore struct {
	UserID int64 `db:"user_id" json:"user_id"`

	// Five scoring-engine dimensions: raw pre-normalization value + 0–100 percentile rank.
	OutputRaw               float64 `db:"output_raw"               json:"output_raw"`
	OutputPercentile        float64 `db:"output_percentile"        json:"output_percentile"`
	CraftRaw                float64 `db:"craft_raw"                json:"craft_raw"`
	CraftPercentile         float64 `db:"craft_percentile"         json:"craft_percentile"`
	InfluenceRaw            float64 `db:"influence_raw"            json:"influence_raw"`
	InfluencePercentile     float64 `db:"influence_percentile"     json:"influence_percentile"`
	CollaborationRaw        float64 `db:"collaboration_raw"        json:"collaboration_raw"`
	CollaborationPercentile float64 `db:"collaboration_percentile" json:"collaboration_percentile"`
	RangeRaw                float64 `db:"range_raw"                json:"range_raw"`
	RangePercentile         float64 `db:"range_percentile"         json:"range_percentile"`

	// Trust is the 0–1 meta-score (never ranked; increases with verified evidence).
	Trust float64 `db:"trust" json:"trust"`

	// Set when scores were last computed.
	GitHubUsername *string    `db:"github_username" json:"github_username,omitempty"`
	ComputedAt     *time.Time `db:"computed_at"     json:"computed_at,omitempty"`

	// Scoring job lifecycle.
	ScoringStatus    *string    `db:"scoring_status"     json:"scoring_status,omitempty"`
	ScoringStartedAt *time.Time `db:"scoring_started_at" json:"scoring_started_at,omitempty"`
	ScoringDoneAt    *time.Time `db:"scoring_done_at"    json:"scoring_done_at,omitempty"`
	ScoringError     *string    `db:"scoring_error"      json:"scoring_error,omitempty"`

	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
