# rss-communicator

The service now supports two database backends:

- `postgres` for multi-process or heavier deployments
- `sqlite` for single-node/self-hosted deployments on a small VPS

Runtime config:

```env
DB_DRIVER=postgres
DB_DSN=postgres://rss:rss@localhost:5432/rss?sslmode=disable
```

SQLite example:

```env
DB_DRIVER=sqlite
DB_DSN=file:rss.db
```

Migrations are split by dialect:

- `migrations/postgres`
- `migrations/sqlite`

Examples:

```bash
make migrate-up-postgres DB_DSN=postgres://rss:rss@localhost:5432/rss?sslmode=disable
make migrate-up-sqlite DB_DSN=file:rss.db
```
