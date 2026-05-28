-- +goose Up
CREATE TABLE user_signal_scores (
    user_id            BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    builder_score      INT NOT NULL DEFAULT 0,
    thinker_score      INT NOT NULL DEFAULT 0,
    executor_score     INT NOT NULL DEFAULT 0,
    collaborator_score INT NOT NULL DEFAULT 0,
    specialist_score   INT NOT NULL DEFAULT 0,
    trusted_score      INT NOT NULL DEFAULT 0,
    total_signal       INT NOT NULL DEFAULT 0,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_signal_scores;
