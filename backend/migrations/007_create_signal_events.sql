-- +goose Up
CREATE TABLE signal_events (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    evidence_item_id      BIGINT    REFERENCES evidence_items(id) ON DELETE SET NULL,
    dimension             TEXT      NOT NULL, -- builder | thinker | executor | collaborator | specialist | trusted
    base_points           INT       NOT NULL DEFAULT 0,
    weight_multiplier     NUMERIC(4,2) NOT NULL DEFAULT 1.00,
    confidence_multiplier NUMERIC(4,2) NOT NULL DEFAULT 1.00,
    final_points          INT       NOT NULL DEFAULT 0,
    explanation           TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_signal_events_user_id          ON signal_events(user_id);
CREATE INDEX idx_signal_events_evidence_item_id ON signal_events(evidence_item_id);

-- +goose Down
DROP TABLE IF EXISTS signal_events;
