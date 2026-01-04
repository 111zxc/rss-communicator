-- +goose Up
CREATE TABLE IF NOT EXISTS contact_telegram_config (
  contact_id   UUID PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  username     TEXT NULL
);

CREATE TABLE IF NOT EXISTS contact_http_config (
  contact_id      UUID PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  method          TEXT NOT NULL DEFAULT 'POST',
  url             TEXT NOT NULL,
  headers_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
  body_template   TEXT NULL,
  auth_type       TEXT NULL,
  auth_config     JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS contact_email_config (
  contact_id   UUID PRIMARY KEY REFERENCES contacts(id) ON DELETE CASCADE,
  format       TEXT NULL
);

-- +goose Down
DROP TABLE IF EXISTS contact_email_config;
DROP TABLE IF EXISTS contact_http_config;
DROP TABLE IF EXISTS contact_telegram_config;
