# rss-communicator

`rss-communicator` - отправляет рсс ленты =)

- `rssd` - HTTP API и рантайм-воркеры
- `tg-bot` - Telegram-бот

## варианты контактов

http telega email

## хранилище

postgre/sqlite

## очередь

inmem/nats

## usage

```bash
make build
make test
make run-rssd
make run-tg
```

### Docker Compose

```bash
make compose-up
make compose-down
```

```bash
make compose-up-sqlite-memory
make compose-up-postgres-memory
make compose-up-sqlite-nats
make compose-up-postgres-nats
```

### Миграции

```bash
make migrate-up-postgres DB_DSN=postgres://rss:rss@localhost:5432/rss?sslmode=disable
make migrate-up-sqlite DB_DSN=file:rss.db
```