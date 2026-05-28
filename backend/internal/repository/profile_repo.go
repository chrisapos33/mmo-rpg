package repository

import (
	"context"
	"database/sql"
	"errors"

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
