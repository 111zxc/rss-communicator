-- +goose Up
ALTER TABLE feeds ADD COLUMN initialized_at DATETIME NULL;

CREATE INDEX IF NOT EXISTS feeds_initialized_idx
  ON feeds(initialized_at);

-- +goose Down
DROP INDEX IF EXISTS feeds_initialized_idx;
ALTER TABLE feeds DROP COLUMN initialized_at;
