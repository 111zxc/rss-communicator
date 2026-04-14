-- +goose Up
CREATE TABLE IF NOT EXISTS subscriptions (
  id          TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  feed_id     TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  contact_id  TEXT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  enabled     INTEGER NOT NULL DEFAULT 1,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_feed_contact_uniq
  ON subscriptions(feed_id, contact_id);

CREATE INDEX IF NOT EXISTS subscriptions_feed_idx ON subscriptions(feed_id);
CREATE INDEX IF NOT EXISTS subscriptions_contact_idx ON subscriptions(contact_id);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
