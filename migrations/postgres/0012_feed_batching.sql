-- +goose Up
ALTER TABLE feeds
  ADD COLUMN IF NOT EXISTS batch_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS batch_window_seconds INTEGER NOT NULL DEFAULT 3600 CHECK (batch_window_seconds >= 60);

CREATE INDEX IF NOT EXISTS feeds_batch_enabled_idx
  ON feeds(batch_enabled);

-- +goose Down
DROP INDEX IF EXISTS feeds_batch_enabled_idx;

ALTER TABLE feeds
  DROP COLUMN IF EXISTS batch_window_seconds,
  DROP COLUMN IF EXISTS batch_enabled;
