-- +goose Up
DO $$ BEGIN
  CREATE TYPE contact_type AS ENUM ('telegram', 'email', 'http');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE contact_status AS ENUM ('pending', 'active', 'disabled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE delivery_status AS ENUM ('pending', 'in_progress', 'sent', 'failed', 'dead');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE outbox_status AS ENUM ('pending', 'published', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- +goose Down
DROP TYPE IF EXISTS outbox_status;
DROP TYPE IF EXISTS delivery_status;
DROP TYPE IF EXISTS contact_status;
DROP TYPE IF EXISTS contact_type;
