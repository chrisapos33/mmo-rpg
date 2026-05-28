package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

type CVRepo struct {
	db *sqlx.DB
}

func NewCVRepo(db *sqlx.DB) *CVRepo {
	return &CVRepo{db: db}
}

func (r *CVRepo) Create(ctx context.Context, userID int64, storagePath, originalName string) (*domain.CVUpload, error) {
	var u domain.CVUpload
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO cv_uploads (user_id, storage_path, original_name, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, storage_path, original_name, status, extracted_data,
		           error_message, created_at, processed_at`,
		userID, storagePath, originalName, domain.CVStatusProcessing,
	).StructScan(&u)
	return &u, err
}

func (r *CVRepo) MarkDone(ctx context.Context, id int64, data *domain.CVData) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = r.db.ExecContext(ctx,
		`UPDATE cv_uploads
		 SET status = $1, extracted_data = $2::jsonb, processed_at = $3
		 WHERE id = $4`,
		domain.CVStatusDone, string(jsonBytes), now, id,
	)
	return err
}

func (r *CVRepo) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE cv_uploads
		 SET status = $1, error_message = $2, processed_at = $3
		 WHERE id = $4`,
		domain.CVStatusFailed, errMsg, now, id,
	)
	return err
}

func (r *CVRepo) LatestByUserID(ctx context.Context, userID int64) (*domain.CVUpload, error) {
	var u domain.CVUpload
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, storage_path, original_name, status, extracted_data,
		        error_message, created_at, processed_at
		 FROM cv_uploads
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
	).StructScan(&u)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}
