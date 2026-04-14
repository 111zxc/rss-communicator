-- +goose Up
ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'direct',
  ADD COLUMN IF NOT EXISTS source_group_id UUID NULL REFERENCES groups(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS subscriptions_feed_contact_uniq;

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_direct_uniq
  ON subscriptions(feed_id, contact_id)
  WHERE source = 'direct';

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_group_uniq
  ON subscriptions(feed_id, contact_id, source_group_id)
  WHERE source = 'group';

CREATE INDEX IF NOT EXISTS subscriptions_source_group_idx
  ON subscriptions(source, source_group_id);

-- +goose Down
DROP INDEX IF EXISTS subscriptions_source_group_idx;
DROP INDEX IF EXISTS subscriptions_group_uniq;
DROP INDEX IF EXISTS subscriptions_direct_uniq;

CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_feed_contact_uniq
  ON subscriptions(feed_id, contact_id);

ALTER TABLE subscriptions
  DROP COLUMN IF EXISTS source_group_id,
  DROP COLUMN IF EXISTS source;
