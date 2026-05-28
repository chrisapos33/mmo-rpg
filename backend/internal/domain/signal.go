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

// Signal dimensions.
const (
	DimBuilder      = "builder"
	DimThinker      = "thinker"
	DimExecutor     = "executor"
	DimCollaborator = "collaborator"
	DimSpecialist   = "specialist"
	DimTrusted      = "trusted"
)

// AllDimensions in the canonical display order.
var AllDimensions = []string{
	DimBuilder, DimExecutor, DimSpecialist, DimTrusted, DimCollaborator, DimThinker,
}

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

// UserSignalScore is the aggregated multi-dimensional signal for a user.
type UserSignalScore struct {
	UserID            int64     `db:"user_id"            json:"user_id"`
	BuilderScore      int       `db:"builder_score"      json:"builder_score"`
	ThinkerScore      int       `db:"thinker_score"      json:"thinker_score"`
	ExecutorScore     int       `db:"executor_score"     json:"executor_score"`
	CollaboratorScore int       `db:"collaborator_score" json:"collaborator_score"`
	SpecialistScore   int       `db:"specialist_score"   json:"specialist_score"`
	TrustedScore      int       `db:"trusted_score"      json:"trusted_score"`
	TotalSignal       int       `db:"total_signal"       json:"total_signal"`
	UpdatedAt         time.Time `db:"updated_at"         json:"updated_at"`
}

// DimScore returns the score for a given dimension string.
func (s *UserSignalScore) DimScore(dim string) int {
	switch dim {
	case DimBuilder:
		return s.BuilderScore
	case DimThinker:
		return s.ThinkerScore
	case DimExecutor:
		return s.ExecutorScore
	case DimCollaborator:
		return s.CollaboratorScore
	case DimSpecialist:
		return s.SpecialistScore
	case DimTrusted:
		return s.TrustedScore
	}
	return 0
}
