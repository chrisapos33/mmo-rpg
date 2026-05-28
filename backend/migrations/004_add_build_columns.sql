-- +goose Up
-- One profile per user — enables ON CONFLICT upsert
ALTER TABLE profiles ADD CONSTRAINT profiles_user_id_key UNIQUE (user_id);

ALTER TABLE profiles
  ADD COLUMN strengths    text[] NOT NULL DEFAULT '{}',
  ADD COLUMN growth_paths text[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_user_id_key;
ALTER TABLE profiles DROP COLUMN IF EXISTS strengths;
ALTER TABLE profiles DROP COLUMN IF EXISTS growth_paths;
