-- +goose Up
CREATE TABLE IF NOT EXISTS items (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  feed_id        UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,

  external_id    TEXT NULL,
  uniq_key       TEXT NOT NULL,

  title          TEXT NOT NULL,
  link           TEXT NOT NULL,
  summary        TEXT NULL,
  author         TEXT NULL,
  published_at   TIMESTAMPTZ NULL,

  raw_json       JSONB NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS items_feed_external_uniq
  ON items(feed_id, external_id)
  WHERE external_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS items_feed_uniqkey_uniq
  ON items(feed_id, uniq_key);

CREATE INDEX IF NOT EXISTS items_feed_created_idx
  ON items(feed_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS items;
