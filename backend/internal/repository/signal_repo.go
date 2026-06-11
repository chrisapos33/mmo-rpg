package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

type SignalRepo struct {
	db *sqlx.DB
}

func NewSignalRepo(db *sqlx.DB) *SignalRepo {
	return &SignalRepo{db: db}
}

// UpsertEvidence creates or updates an evidence item identified by (user_id, source_type, source_key).
func (r *SignalRepo) UpsertEvidence(ctx context.Context, e *domain.EvidenceItem) (*domain.EvidenceItem, error) {
	var out domain.EvidenceItem
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO evidence_items
		  (user_id, source_type, source_key, artifact_url, title, description,
		   metadata_json, verification_status, verification_confidence)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (user_id, source_type, source_key) DO UPDATE SET
		  artifact_url            = EXCLUDED.artifact_url,
		  title                   = EXCLUDED.title,
		  description             = EXCLUDED.description,
		  metadata_json           = EXCLUDED.metadata_json,
		  verification_status     = EXCLUDED.verification_status,
		  verification_confidence = EXCLUDED.verification_confidence,
		  updated_at              = NOW()
		RETURNING
		  id, user_id, source_type, source_key, artifact_url, title, description,
		  metadata_json, verification_status, verification_confidence, created_at, updated_at`,
		e.UserID, e.SourceType, e.SourceKey, e.ArtifactURL, e.Title, e.Description,
		e.MetadataJSON, e.VerificationStatus, e.VerificationConfidence,
	).StructScan(&out)
	return &out, err
}

// ReplaceSignalEvents deletes all events for an evidence item then bulk-inserts the new set.
func (r *SignalRepo) ReplaceSignalEvents(ctx context.Context, evidenceItemID int64, events []*domain.SignalEvent) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM signal_events WHERE evidence_item_id = $1`, evidenceItemID,
	); err != nil {
		return err
	}

	for _, ev := range events {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO signal_events
			  (user_id, evidence_item_id, dimension, base_points,
			   weight_multiplier, confidence_multiplier, final_points, explanation)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			ev.UserID, ev.EvidenceItemID, ev.Dimension, ev.BasePoints,
			ev.WeightMultiplier, ev.ConfidenceMultiplier, ev.FinalPoints, ev.Explanation,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetScores returns the user's current signal scores, or ErrNotFound.
func (r *SignalRepo) GetScores(ctx context.Context, userID int64) (*domain.UserSignalScore, error) {
	var score domain.UserSignalScore
	err := r.db.QueryRowxContext(ctx, `
		SELECT user_id,
		       output_raw, output_percentile,
		       craft_raw, craft_percentile,
		       influence_raw, influence_percentile,
		       collaboration_raw, collaboration_percentile,
		       range_raw, range_percentile,
		       trust,
		       github_username, computed_at,
		       scoring_status, scoring_started_at, scoring_done_at, scoring_error,
		       updated_at
		FROM user_signal_scores
		WHERE user_id = $1`,
		userID,
	).StructScan(&score)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &score, err
}

// StartScoringJob atomically claims the scoring slot for a user.
// Creates a new row or updates an existing one to status='running'.
// Returns false without error when a non-stale run is already in progress.
// A run is considered stale after 6 minutes (matches the 5-minute ingestion timeout
// plus a 1-minute buffer), so a crashed server can never orphan-lock a user.
func (r *SignalRepo) StartScoringJob(ctx context.Context, userID int64) (bool, error) {
	var id int64
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO user_signal_scores (user_id, scoring_status, scoring_started_at, updated_at)
		VALUES ($1, 'running', NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
		    scoring_status     = 'running',
		    scoring_started_at = NOW(),
		    scoring_done_at    = NULL,
		    scoring_error      = NULL,
		    updated_at         = NOW()
		WHERE user_signal_scores.scoring_status IS DISTINCT FROM 'running'
		   OR user_signal_scores.scoring_started_at < NOW() - INTERVAL '6 minutes'
		RETURNING user_id`,
		userID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // blocked by an active, non-stale run
	}
	return err == nil, err
}

// SaveGitHubScores writes all five dimension scores, trust, and metadata, then
// marks the scoring job done. Called by the background runScoring goroutine on success.
func (r *SignalRepo) SaveGitHubScores(ctx context.Context, userID int64, username string, scores scoring.Scores) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_signal_scores SET
		    output_raw               = $2,
		    output_percentile        = $3,
		    craft_raw                = $4,
		    craft_percentile         = $5,
		    influence_raw            = $6,
		    influence_percentile     = $7,
		    collaboration_raw        = $8,
		    collaboration_percentile = $9,
		    range_raw                = $10,
		    range_percentile         = $11,
		    trust                    = $12,
		    github_username          = $13,
		    computed_at              = $14,
		    scoring_status           = 'done',
		    scoring_done_at          = NOW(),
		    scoring_error            = NULL,
		    updated_at               = NOW()
		WHERE user_id = $1`,
		userID,
		scores.Output.Raw, scores.Output.Percentile,
		scores.Craft.Raw, scores.Craft.Percentile,
		scores.Influence.Raw, scores.Influence.Percentile,
		scores.Collaboration.Raw, scores.Collaboration.Percentile,
		scores.Range.Raw, scores.Range.Percentile,
		scores.Trust,
		username,
		scores.ComputedAt,
	)
	return err
}

// FailScoringJob marks the scoring job as failed with a reason string.
func (r *SignalRepo) FailScoringJob(ctx context.Context, userID int64, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_signal_scores SET
		    scoring_status  = 'failed',
		    scoring_done_at = NOW(),
		    scoring_error   = $2,
		    updated_at      = NOW()
		WHERE user_id = $1`,
		userID, reason,
	)
	return err
}

// DeleteEvidence removes an evidence item (verifying ownership) along with its
// signal_events in a single transaction.
func (r *SignalRepo) DeleteEvidence(ctx context.Context, userID, evidenceID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM signal_events WHERE evidence_item_id = $1 AND user_id = $2`,
		evidenceID, userID,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`DELETE FROM evidence_items WHERE id = $1 AND user_id = $2`,
		evidenceID, userID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

// ListEvidence returns all evidence items for a user, newest first.
func (r *SignalRepo) ListEvidence(ctx context.Context, userID int64) ([]*domain.EvidenceItem, error) {
	var items []*domain.EvidenceItem
	err := r.db.SelectContext(ctx, &items, `
		SELECT id, user_id, source_type, source_key, artifact_url, title, description,
		       metadata_json, verification_status, verification_confidence, created_at, updated_at
		FROM evidence_items
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID,
	)
	return items, err
}
