-- +goose Up
CREATE TABLE cv_uploads (
    id             BIGSERIAL   PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_path   TEXT        NOT NULL,
    original_name  TEXT        NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'processing',
    extracted_data JSONB,
    error_message  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at   TIMESTAMPTZ
);

CREATE INDEX idx_cv_uploads_user_id ON cv_uploads(user_id);

-- +goose Down
DROP TABLE cv_uploads;
