-- +goose Up

-- Replace the old free-form dimension columns (builder/thinker/executor/collaborator/
-- specialist/trusted/total_signal) with the scoring engine's five-dimension taxonomy.
-- Also add scoring job status columns alongside the score columns so the frontend can
-- poll a single row instead of a separate table.
--
-- Pre-launch dev: no real user data. Hard cut — no backfill.

ALTER TABLE user_signal_scores
    DROP COLUMN builder_score,
    DROP COLUMN thinker_score,
    DROP COLUMN executor_score,
    DROP COLUMN collaborator_score,
    DROP COLUMN specialist_score,
    DROP COLUMN trusted_score,
    DROP COLUMN total_signal,
    ADD COLUMN output_raw               DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN output_percentile        DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN craft_raw                DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN craft_percentile         DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN influence_raw            DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN influence_percentile     DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN collaboration_raw        DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN collaboration_percentile DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN range_raw                DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN range_percentile         DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN trust                    DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN github_username          TEXT,
    ADD COLUMN computed_at              TIMESTAMPTZ,
    -- Scoring job lifecycle — mirrors the cv_uploads.status pattern.
    ADD COLUMN scoring_status           TEXT,
    ADD COLUMN scoring_started_at       TIMESTAMPTZ,
    ADD COLUMN scoring_done_at          TIMESTAMPTZ,
    ADD COLUMN scoring_error            TEXT;

-- +goose Down

ALTER TABLE user_signal_scores
    ADD COLUMN builder_score      INT NOT NULL DEFAULT 0,
    ADD COLUMN thinker_score      INT NOT NULL DEFAULT 0,
    ADD COLUMN executor_score     INT NOT NULL DEFAULT 0,
    ADD COLUMN collaborator_score INT NOT NULL DEFAULT 0,
    ADD COLUMN specialist_score   INT NOT NULL DEFAULT 0,
    ADD COLUMN trusted_score      INT NOT NULL DEFAULT 0,
    ADD COLUMN total_signal       INT NOT NULL DEFAULT 0,
    DROP COLUMN output_raw,
    DROP COLUMN output_percentile,
    DROP COLUMN craft_raw,
    DROP COLUMN craft_percentile,
    DROP COLUMN influence_raw,
    DROP COLUMN influence_percentile,
    DROP COLUMN collaboration_raw,
    DROP COLUMN collaboration_percentile,
    DROP COLUMN range_raw,
    DROP COLUMN range_percentile,
    DROP COLUMN trust,
    DROP COLUMN github_username,
    DROP COLUMN computed_at,
    DROP COLUMN scoring_status,
    DROP COLUMN scoring_started_at,
    DROP COLUMN scoring_done_at,
    DROP COLUMN scoring_error;
