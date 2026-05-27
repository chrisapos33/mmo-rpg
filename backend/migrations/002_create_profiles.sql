-- +goose Up
CREATE TABLE profiles (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username        TEXT        UNIQUE,
    display_name    TEXT,
    class           TEXT,
    subclass        TEXT,
    headline        TEXT,
    summary         TEXT,
    avatar_url      TEXT,
    signal_score    INT         NOT NULL DEFAULT 0,
    xp              INT         NOT NULL DEFAULT 0,
    is_published    BOOLEAN     NOT NULL DEFAULT FALSE,
    onboarding_step TEXT        NOT NULL DEFAULT 'upload',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id  ON profiles(user_id);
CREATE INDEX idx_profiles_username ON profiles(username);

-- +goose Down
DROP TABLE profiles;
