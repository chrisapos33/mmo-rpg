-- +goose Up
CREATE TABLE evidence_items (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type             TEXT      NOT NULL, -- github | blog | portfolio | community | manual | linkedin | other
    source_key              TEXT      NOT NULL, -- disambiguates multiple items of same type (github_user_id, url, etc.)
    artifact_url            TEXT,
    title                   TEXT      NOT NULL,
    description             TEXT,
    metadata_json           JSONB,
    verification_status     TEXT      NOT NULL DEFAULT 'unverified', -- unverified | url_verified | platform_verified | peer_verified | admin_verified
    verification_confidence NUMERIC(4,3) NOT NULL DEFAULT 0.000,    -- 0.000 – 1.000
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, source_type, source_key)
);

CREATE INDEX idx_evidence_items_user_id ON evidence_items(user_id);

-- +goose Down
DROP TABLE IF EXISTS evidence_items;
