package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DB struct {
		DSN string
	}
	Telegram struct {
		BotToken string
	}
	Queue struct {
		Driver  string
		NatsURL string
	}
	LogLevel slog.Level
}

func MustLoad() Config {
	var c Config
	c.DB.DSN = Getenv("DB_DSN", "postgres://rss:rss@localhost:5432/rss?sslmode=disable")
	c.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	c.Queue.Driver = Getenv("QUEUE_DRIVER", "memory")
	c.Queue.NatsURL = Getenv("NATS_URL", "nats://localhost:4222")

	c.LogLevel = parseLevel(Getenv("LOG_LEVEL", "info"))
	return c
}

func Getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
