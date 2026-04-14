-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
  id           TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  name         TEXT NOT NULL,
  description  TEXT NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS groups_name_uniq ON groups(name);

CREATE TABLE IF NOT EXISTS group_contacts (
  group_id     TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  contact_id   TEXT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (group_id, contact_id)
);

CREATE INDEX IF NOT EXISTS group_contacts_contact_idx ON group_contacts(contact_id);

CREATE TABLE IF NOT EXISTS group_feeds (
  group_id     TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  feed_id      TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (group_id, feed_id)
);

CREATE INDEX IF NOT EXISTS group_feeds_feed_idx ON group_feeds(feed_id);

CREATE TABLE IF NOT EXISTS registration_codes (
  id           TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  code         TEXT NOT NULL,
  name         TEXT NOT NULL,
  description  TEXT NULL,
  enabled      INTEGER NOT NULL DEFAULT 1,
  max_uses     INTEGER NULL CHECK (max_uses IS NULL OR max_uses > 0),
  use_count    INTEGER NOT NULL DEFAULT 0,
  expires_at   DATETIME NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS registration_codes_code_uniq ON registration_codes(code);

CREATE TABLE IF NOT EXISTS registration_code_groups (
  registration_code_id TEXT NOT NULL REFERENCES registration_codes(id) ON DELETE CASCADE,
  group_id             TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (registration_code_id, group_id)
);

CREATE INDEX IF NOT EXISTS registration_code_groups_group_idx ON registration_code_groups(group_id);

-- +goose Down
DROP TABLE IF EXISTS registration_code_groups;
DROP TABLE IF EXISTS registration_codes;
DROP TABLE IF EXISTS group_feeds;
DROP TABLE IF EXISTS group_contacts;
DROP TABLE IF EXISTS groups;
