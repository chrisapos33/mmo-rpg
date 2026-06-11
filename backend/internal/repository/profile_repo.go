package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

type ProfileRepo struct {
	db *sqlx.DB
}

func NewProfileRepo(db *sqlx.DB) *ProfileRepo {
	return &ProfileRepo{db: db}
}

// UpsertBuild creates or updates the profile with the generated build data.
func (r *ProfileRepo) UpsertBuild(ctx context.Context, userID int64, build *domain.BuildData) (*domain.Profile, error) {
	var p domain.Profile
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO profiles
		  (user_id, class, subclass, headline, summary, strengths, growth_paths, onboarding_step)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'reveal')
		ON CONFLICT (user_id) DO UPDATE SET
		  class           = EXCLUDED.class,
		  subclass        = EXCLUDED.subclass,
		  headline        = EXCLUDED.headline,
		  summary         = EXCLUDED.summary,
		  strengths       = EXCLUDED.strengths,
		  growth_paths    = EXCLUDED.growth_paths,
		  onboarding_step = 'reveal',
		  updated_at      = NOW()
		RETURNING
		  id, user_id, username, display_name, class, subclass, headline, summary,
		  avatar_url, signal_score, xp, is_published, onboarding_step,
		  strengths, growth_paths, created_at, updated_at`,
		userID,
		build.Class,
		build.Subclass,
		build.Headline,
		build.Summary,
		pq.Array(build.Strengths),
		pq.Array(build.GrowthPaths),
	).StructScan(&p)
	return &p, err
}

// UpdateSignalScore recalculates and stores the signal score for a user's profile.
func (r *ProfileRepo) UpdateSignalScore(ctx context.Context, userID int64, score int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE profiles SET signal_score = $1, updated_at = NOW() WHERE user_id = $2`,
		score, userID,
	)
	return err
}

// Publish sets is_published = true on a user's profile.
func (r *ProfileRepo) Publish(ctx context.Context, userID int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE profiles SET is_published = true, updated_at = NOW() WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// FindPublicByUserID returns a published profile, or ErrNotFound.
func (r *ProfileRepo) FindPublicByUserID(ctx context.Context, userID int64) (*domain.Profile, error) {
	var p domain.Profile
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, user_id, username, display_name, class, subclass, headline, summary,
		       avatar_url, signal_score, xp, is_published, onboarding_step,
		       strengths, growth_paths, created_at, updated_at
		FROM profiles
		WHERE user_id = $1 AND is_published = true`,
		userID,
	).StructScan(&p)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

// ListPublished returns published profile cards joined with signal scores and GitHub data.
// class filters by exact class name when non-empty. sort is "signal" or "recent".
func (r *ProfileRepo) ListPublished(ctx context.Context, class, sort string, limit, offset int) ([]*domain.ExploreEntry, error) {
	q := `
		SELECT
		  p.user_id,
		  p.class,
		  p.subclass,
		  p.headline,
		  COALESCE(s.trust, 0)           AS trust,
		  gc.github_username,
		  COALESCE(gc.top_languages, '{}') AS top_languages,
		  p.updated_at
		FROM profiles p
		LEFT JOIN user_signal_scores s  ON s.user_id  = p.user_id
		LEFT JOIN github_connections gc ON gc.user_id = p.user_id
		WHERE p.is_published = true`

	args := []any{}

	if class != "" {
		args = append(args, class)
		q += fmt.Sprintf(" AND p.class = $%d", len(args))
	}

	if sort == "signal" {
		q += " ORDER BY trust DESC, p.updated_at DESC"
	} else {
		q += " ORDER BY p.updated_at DESC"
	}

	args = append(args, limit, offset)
	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	var entries []*domain.ExploreEntry
	err := r.db.SelectContext(ctx, &entries, q, args...)
	return entries, err
}

// FindByUserID returns the profile for a user, or ErrNotFound.
func (r *ProfileRepo) FindByUserID(ctx context.Context, userID int64) (*domain.Profile, error) {
	var p domain.Profile
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, user_id, username, display_name, class, subclass, headline, summary,
		       avatar_url, signal_score, xp, is_published, onboarding_step,
		       strengths, growth_paths, created_at, updated_at
		FROM profiles
		WHERE user_id = $1`,
		userID,
	).StructScan(&p)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}
