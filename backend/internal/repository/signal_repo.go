package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
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

// RecomputeScores aggregates all signal_events for a user and upserts user_signal_scores.
func (r *SignalRepo) RecomputeScores(ctx context.Context, userID int64) (*domain.UserSignalScore, error) {
	var score domain.UserSignalScore
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO user_signal_scores
		  (user_id, builder_score, thinker_score, executor_score,
		   collaborator_score, specialist_score, trusted_score, total_signal, updated_at)
		SELECT
		  $1,
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'builder'),      0),
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'thinker'),      0),
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'executor'),     0),
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'collaborator'), 0),
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'specialist'),   0),
		  COALESCE(SUM(final_points) FILTER (WHERE dimension = 'trusted'),      0),
		  COALESCE(SUM(final_points), 0),
		  NOW()
		FROM signal_events
		WHERE user_id = $1
		ON CONFLICT (user_id) DO UPDATE SET
		  builder_score      = EXCLUDED.builder_score,
		  thinker_score      = EXCLUDED.thinker_score,
		  executor_score     = EXCLUDED.executor_score,
		  collaborator_score = EXCLUDED.collaborator_score,
		  specialist_score   = EXCLUDED.specialist_score,
		  trusted_score      = EXCLUDED.trusted_score,
		  total_signal       = EXCLUDED.total_signal,
		  updated_at         = NOW()
		RETURNING
		  user_id, builder_score, thinker_score, executor_score,
		  collaborator_score, specialist_score, trusted_score, total_signal, updated_at`,
		userID,
	).StructScan(&score)
	return &score, err
}

// GetScores returns the user's current signal scores, or ErrNotFound.
func (r *SignalRepo) GetScores(ctx context.Context, userID int64) (*domain.UserSignalScore, error) {
	var score domain.UserSignalScore
	err := r.db.QueryRowxContext(ctx, `
		SELECT user_id, builder_score, thinker_score, executor_score,
		       collaborator_score, specialist_score, trusted_score, total_signal, updated_at
		FROM user_signal_scores WHERE user_id = $1`,
		userID,
	).StructScan(&score)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &score, err
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
