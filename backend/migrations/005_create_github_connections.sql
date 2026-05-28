-- +goose Up
CREATE TABLE github_connections (
    id                 BIGSERIAL   PRIMARY KEY,
    user_id            BIGINT      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    github_username    TEXT        NOT NULL,
    github_user_id     BIGINT      NOT NULL,
    access_token       TEXT        NOT NULL,
    avatar_url         TEXT,
    repo_count         INT         NOT NULL DEFAULT 0,
    star_count         INT         NOT NULL DEFAULT 0,
    followers          INT         NOT NULL DEFAULT 0,
    top_languages      text[]      NOT NULL DEFAULT '{}',
    contribution_score INT         NOT NULL DEFAULT 0,
    synced_at          TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE github_connections;
