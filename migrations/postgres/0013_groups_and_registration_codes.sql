-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name         TEXT NOT NULL,
  description  TEXT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS groups_name_uniq ON groups(name);

CREATE TABLE IF NOT EXISTS group_contacts (
  group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  contact_id   UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, contact_id)
);

CREATE INDEX IF NOT EXISTS group_contacts_contact_idx ON group_contacts(contact_id);

CREATE TABLE IF NOT EXISTS group_feeds (
  group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  feed_id      UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, feed_id)
);

CREATE INDEX IF NOT EXISTS group_feeds_feed_idx ON group_feeds(feed_id);

CREATE TABLE IF NOT EXISTS registration_codes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         TEXT NOT NULL,
  name         TEXT NOT NULL,
  description  TEXT NULL,
  enabled      BOOLEAN NOT NULL DEFAULT TRUE,
  max_uses     INTEGER NULL CHECK (max_uses IS NULL OR max_uses > 0),
  use_count    INTEGER NOT NULL DEFAULT 0,
  expires_at   TIMESTAMPTZ NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS registration_codes_code_uniq ON registration_codes(code);

CREATE TABLE IF NOT EXISTS registration_code_groups (
  registration_code_id UUID NOT NULL REFERENCES registration_codes(id) ON DELETE CASCADE,
  group_id             UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (registration_code_id, group_id)
);

CREATE INDEX IF NOT EXISTS registration_code_groups_group_idx ON registration_code_groups(group_id);

-- +goose Down
DROP TABLE IF EXISTS registration_code_groups;
DROP TABLE IF EXISTS registration_codes;
DROP TABLE IF EXISTS group_feeds;
DROP TABLE IF EXISTS group_contacts;
DROP TABLE IF EXISTS groups;
