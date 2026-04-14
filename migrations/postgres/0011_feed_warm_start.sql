-- +goose Up
ALTER TABLE feeds
  ADD COLUMN IF NOT EXISTS initialized_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS feeds_initialized_idx
  ON feeds(initialized_at);

-- +goose Down
DROP INDEX IF EXISTS feeds_initialized_idx;

ALTER TABLE feeds
  DROP COLUMN IF EXISTS initialized_at;
