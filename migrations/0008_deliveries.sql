-- +goose Up
CREATE TABLE IF NOT EXISTS deliveries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id         UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  contact_id      UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,

  status          delivery_status NOT NULL DEFAULT 'pending',
  attempt_count   INTEGER NOT NULL DEFAULT 0,

  last_error      TEXT NULL,
  last_attempt_at TIMESTAMPTZ NULL,
  sent_at         TIMESTAMPTZ NULL,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS deliveries_contact_item_uniq
  ON deliveries(contact_id, item_id);

CREATE INDEX IF NOT EXISTS deliveries_status_idx
  ON deliveries(status, created_at);

-- +goose Down
DROP TABLE IF EXISTS deliveries;
