package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

type GitHubRepo struct {
	db *sqlx.DB
}

func NewGitHubRepo(db *sqlx.DB) *GitHubRepo {
	return &GitHubRepo{db: db}
}

// Upsert creates or updates a GitHub connection and its aggregated stats.
func (r *GitHubRepo) Upsert(ctx context.Context, conn *domain.GitHubConnection) (*domain.GitHubConnection, error) {
	now := time.Now()
	var out domain.GitHubConnection
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO github_connections
		  (user_id, github_username, github_user_id, access_token, avatar_url,
		   repo_count, star_count, followers, top_languages, contribution_score, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (user_id) DO UPDATE SET
		  github_username    = EXCLUDED.github_username,
		  github_user_id     = EXCLUDED.github_user_id,
		  access_token       = EXCLUDED.access_token,
		  avatar_url         = EXCLUDED.avatar_url,
		  repo_count         = EXCLUDED.repo_count,
		  star_count         = EXCLUDED.star_count,
		  followers          = EXCLUDED.followers,
		  top_languages      = EXCLUDED.top_languages,
		  contribution_score = EXCLUDED.contribution_score,
		  synced_at          = EXCLUDED.synced_at,
		  updated_at         = NOW()
		RETURNING
		  id, user_id, github_username, github_user_id, access_token, avatar_url,
		  repo_count, star_count, followers, top_languages, contribution_score,
		  synced_at, created_at, updated_at`,
		conn.UserID, conn.GitHubUsername, conn.GitHubUserID, conn.AccessToken,
		conn.AvatarURL, conn.RepoCount, conn.StarCount, conn.Followers,
		pq.Array(conn.TopLanguages), conn.ContributionScore, now,
	).StructScan(&out)
	return &out, err
}

// FindByUserID returns the GitHub connection for a user.
func (r *GitHubRepo) FindByUserID(ctx context.Context, userID int64) (*domain.GitHubConnection, error) {
	var conn domain.GitHubConnection
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, user_id, github_username, github_user_id, access_token, avatar_url,
		       repo_count, star_count, followers, top_languages, contribution_score,
		       synced_at, created_at, updated_at
		FROM github_connections
		WHERE user_id = $1`,
		userID,
	).StructScan(&conn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &conn, err
}
