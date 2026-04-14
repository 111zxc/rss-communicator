-- +goose Up
CREATE TABLE IF NOT EXISTS subscriptions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  feed_id     UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  enabled     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_feed_contact_uniq
  ON subscriptions(feed_id, contact_id);

CREATE INDEX IF NOT EXISTS subscriptions_feed_idx ON subscriptions(feed_id);
CREATE INDEX IF NOT EXISTS subscriptions_contact_idx ON subscriptions(contact_id);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
