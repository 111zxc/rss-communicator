-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic         TEXT NOT NULL,
  payload       JSONB NOT NULL,
  status        outbox_status NOT NULL DEFAULT 'pending',

  available_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  attempt_count INTEGER NOT NULL DEFAULT 0,
  last_error    TEXT NULL,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS outbox_pending_idx
  ON outbox(status, available_at, created_at);

-- +goose Down
DROP TABLE IF EXISTS outbox;
