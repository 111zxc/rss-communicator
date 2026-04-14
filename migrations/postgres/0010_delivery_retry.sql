-- +goose Up
ALTER TABLE deliveries
  ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS deliveries_retry_idx
  ON deliveries(status, next_retry_at)
  WHERE status IN ('pending', 'failed');

-- +goose Down
DROP INDEX IF EXISTS deliveries_retry_idx;

ALTER TABLE deliveries
  DROP COLUMN IF EXISTS next_retry_at;
