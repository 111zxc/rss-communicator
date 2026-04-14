-- +goose Up
CREATE TABLE IF NOT EXISTS contacts (
  id           TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  type         TEXT NOT NULL CHECK (type IN ('telegram', 'email', 'http')),
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'disabled')),
  value        TEXT NOT NULL,
  display_name TEXT NULL,
  verified_at  DATETIME NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS contacts_type_value_uniq ON contacts(type, value);
CREATE INDEX IF NOT EXISTS contacts_status_idx ON contacts(status);

-- +goose Down
DROP TABLE IF EXISTS contacts;
