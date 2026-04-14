-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
  id            TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  topic         TEXT NOT NULL,
  payload       TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'published', 'failed')),
  available_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  last_error    TEXT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS outbox_pending_idx
  ON outbox(status, available_at, created_at);

-- +goose Down
DROP TABLE IF EXISTS outbox;
