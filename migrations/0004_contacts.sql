-- +goose Up
CREATE TABLE IF NOT EXISTS contacts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type         contact_type NOT NULL,
  status       contact_status NOT NULL DEFAULT 'pending',
  value        TEXT NOT NULL,
  display_name TEXT NULL,
  verified_at  TIMESTAMPTZ NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS contacts_type_value_uniq ON contacts(type, value);
CREATE INDEX IF NOT EXISTS contacts_status_idx ON contacts(status);

-- +goose Down
DROP TABLE IF EXISTS contacts;
