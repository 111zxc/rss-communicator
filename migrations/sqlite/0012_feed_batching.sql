-- +goose Up
ALTER TABLE feeds ADD COLUMN batch_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE feeds ADD COLUMN batch_window_seconds INTEGER NOT NULL DEFAULT 3600 CHECK (batch_window_seconds >= 60);

CREATE INDEX IF NOT EXISTS feeds_batch_enabled_idx
  ON feeds(batch_enabled);

-- +goose Down
DROP INDEX IF EXISTS feeds_batch_enabled_idx;
ALTER TABLE feeds DROP COLUMN batch_window_seconds;
ALTER TABLE feeds DROP COLUMN batch_enabled;
