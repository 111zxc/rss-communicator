-- +goose Up
CREATE TABLE IF NOT EXISTS deliveries (
  id              TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
  item_id         TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  contact_id      TEXT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'sent', 'failed', 'dead')),
  attempt_count   INTEGER NOT NULL DEFAULT 0,
  last_error      TEXT NULL,
  last_attempt_at DATETIME NULL,
  sent_at         DATETIME NULL,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS deliveries_contact_item_uniq
  ON deliveries(contact_id, item_id);

CREATE INDEX IF NOT EXISTS deliveries_status_idx
  ON deliveries(status, created_at);

-- +goose Down
DROP TABLE IF EXISTS deliveries;
