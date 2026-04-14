-- +goose Up
CREATE TABLE IF NOT EXISTS contact_telegram_config (
  contact_id   TEXT PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  username     TEXT NULL
);

CREATE TABLE IF NOT EXISTS contact_http_config (
  contact_id      TEXT PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  method          TEXT NOT NULL DEFAULT 'POST',
  url             TEXT NOT NULL,
  headers_json    TEXT NOT NULL DEFAULT '{}',
  body_template   TEXT NULL,
  auth_type       TEXT NULL,
  auth_config     TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS contact_email_config (
  contact_id   TEXT PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  format       TEXT NULL
);

-- +goose Down
DROP TABLE IF EXISTS contact_email_config;
DROP TABLE IF EXISTS contact_http_config;
DROP TABLE IF EXISTS contact_telegram_config;
