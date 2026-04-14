# rss-communicator

The service now supports two database backends:

- `postgres` for multi-process or heavier deployments
- `sqlite` for single-node/self-hosted deployments on a small VPS

Queue backends:

- `memory` for single-process/in-memory runtime jobs
- `nats` for durable broker-backed job transport via JetStream

Runtime config:

```env
DB_DRIVER=postgres
DB_DSN=postgres://rss:rss@localhost:5432/rss?sslmode=disable
QUEUE_DRIVER=memory
```

SQLite example:

```env
DB_DRIVER=sqlite
DB_DSN=file:rss.db
```

NATS example:

```env
QUEUE_DRIVER=nats
NATS_URL=nats://127.0.0.1:4222
NATS_STREAM=RSS_COMMUNICATOR
NATS_SUBJECT_ROOT=rss
```

Jobs are now published durably through the database `outbox` and then forwarded to the selected queue backend. Delivery semantics are `at-least-once`.

Migrations are split by dialect:

- `migrations/postgres`
- `migrations/sqlite`

Examples:

```bash
make migrate-up-postgres DB_DSN=postgres://rss:rss@localhost:5432/rss?sslmode=disable
make migrate-up-sqlite DB_DSN=file:rss.db
```
