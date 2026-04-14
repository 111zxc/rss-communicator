-- +goose Up
CREATE TABLE IF NOT EXISTS feeds (
  id                   TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  url                  TEXT NOT NULL,
  name                 TEXT NOT NULL,
  enabled              INTEGER NOT NULL DEFAULT 1,
  interval_seconds     INTEGER NOT NULL DEFAULT 300 CHECK (interval_seconds >= 10),
  etag                 TEXT NULL,
  last_modified        TEXT NULL,
  last_fetch_at        DATETIME NULL,
  next_fetch_at        DATETIME NULL,
  last_error           TEXT NULL,
  error_count          INTEGER NOT NULL DEFAULT 0,
  created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS feeds_url_uniq ON feeds (url);
CREATE INDEX IF NOT EXISTS feeds_next_fetch_idx ON feeds (enabled, next_fetch_at);

-- +goose Down
DROP TABLE IF EXISTS feeds;
