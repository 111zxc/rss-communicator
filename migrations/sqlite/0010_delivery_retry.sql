-- +goose Up
ALTER TABLE deliveries ADD COLUMN next_retry_at DATETIME NULL;

CREATE INDEX IF NOT EXISTS deliveries_retry_idx
  ON deliveries(status, next_retry_at)
  WHERE status IN ('pending', 'failed');

-- +goose Down
DROP INDEX IF EXISTS deliveries_retry_idx;
ALTER TABLE deliveries DROP COLUMN next_retry_at;
