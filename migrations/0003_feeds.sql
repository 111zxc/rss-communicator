-- +goose Up
CREATE TABLE IF NOT EXISTS feeds (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url              TEXT NOT NULL,
  name             TEXT NOT NULL,
  enabled          BOOLEAN NOT NULL DEFAULT TRUE,

  interval_seconds INTEGER NOT NULL DEFAULT 300 CHECK (interval_seconds >= 10),

  etag             TEXT NULL,
  last_modified    TEXT NULL,

  last_fetch_at    TIMESTAMPTZ NULL,
  next_fetch_at    TIMESTAMPTZ NULL,

  last_error       TEXT NULL,
  error_count      INTEGER NOT NULL DEFAULT 0,

  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS feeds_url_uniq ON feeds (url);
CREATE INDEX IF NOT EXISTS feeds_next_fetch_idx ON feeds (enabled, next_fetch_at);

-- +goose Down
DROP TABLE IF EXISTS feeds;
