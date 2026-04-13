package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type RSSDConfig struct {
	ScheduleTick time.Duration
	FetchBatch   int
	FetchTimeout time.Duration

	DeliveryWorkers int
	TelegramRPS     float64
	TelegramBurst   int
	HTTPTimeout     time.Duration

	RetryScheduleTick time.Duration
	RetryBatch        int
	RetryBase         time.Duration
	RetryMax          time.Duration
	RetryMaxAttempts  int
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type EmailConfig struct {
	Address     string
	DisplayName string

	IMAPHost         string
	IMAPPort         int
	IMAPUsername     string
	IMAPPassword     string
	IMAPMailbox      string
	IMAPPollInterval time.Duration

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

type Config struct {
	DB       struct{ DSN string }
	Telegram struct{ BotToken string }
	Email    EmailConfig
	Log      struct {
		Level     slog.Level
		Format    string
		AddSource bool
	}

	HTTP HTTPConfig
	RSSD RSSDConfig
}

func MustLoad() Config {
	var cfg Config
	cfg.DB.DSN = Getenv("DB_DSN", "postgres://rss:rss@localhost:5432/rss?sslmode=disable")
	cfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.Email.Address = Getenv("EMAIL_ADDRESS", "")
	cfg.Email.DisplayName = Getenv("EMAIL_DISPLAY_NAME", "")
	cfg.Email.IMAPHost = Getenv("EMAIL_IMAP_HOST", "")
	cfg.Email.IMAPPort = MustInt(Getenv("EMAIL_IMAP_PORT", "993"))
	cfg.Email.IMAPUsername = Getenv("EMAIL_IMAP_USERNAME", "")
	cfg.Email.IMAPPassword = Getenv("EMAIL_IMAP_PASSWORD", "")
	cfg.Email.IMAPMailbox = Getenv("EMAIL_IMAP_MAILBOX", "INBOX")
	cfg.Email.IMAPPollInterval = MustDuration(Getenv("EMAIL_IMAP_POLL_INTERVAL", "30s"))
	cfg.Email.SMTPHost = Getenv("EMAIL_SMTP_HOST", "")
	cfg.Email.SMTPPort = MustInt(Getenv("EMAIL_SMTP_PORT", "465"))
	cfg.Email.SMTPUsername = Getenv("EMAIL_SMTP_USERNAME", "")
	cfg.Email.SMTPPassword = Getenv("EMAIL_SMTP_PASSWORD", "")

	cfg.Log.Level = parseLevel(Getenv("LOG_LEVEL", "info"))
	cfg.Log.Format = Getenv("LOG_FORMAT", "json")
	cfg.Log.AddSource = MustBool(Getenv("LOG_ADD_SOURCE", "false"))

	cfg.HTTP.Addr = Getenv("HTTP_ADDR", ":8080")
	cfg.HTTP.ReadTimeout = MustDuration(Getenv("HTTP_READ_TIMEOUT", "5s"))
	cfg.HTTP.WriteTimeout = MustDuration(Getenv("HTTP_WRITE_TIMEOUT", "10s"))
	cfg.HTTP.IdleTimeout = MustDuration(Getenv("HTTP_IDLE_TIMEOUT", "60s"))

	cfg.RSSD.ScheduleTick = MustDuration(Getenv("RSS_SCHEDULE_TICK", "5s"))
	cfg.RSSD.FetchBatch = MustInt(Getenv("RSS_FETCH_BATCH", "50"))
	cfg.RSSD.FetchTimeout = MustDuration(Getenv("FETCH_TIMEOUT", "20s"))

	cfg.RSSD.DeliveryWorkers = MustInt(Getenv("DELIVERY_WORKERS", "4"))
	cfg.RSSD.TelegramRPS = MustFloat(Getenv("TELEGRAM_RPS", "5"))
	cfg.RSSD.TelegramBurst = MustInt(Getenv("TELEGRAM_BURST", "10"))
	cfg.RSSD.HTTPTimeout = MustDuration(Getenv("DELIVERY_HTTP_TIMEOUT", "10s"))

	cfg.RSSD.RetryScheduleTick = MustDuration(Getenv("RETRY_SCHEDULE_TICK", "5s"))
	cfg.RSSD.RetryBatch = MustInt(Getenv("RETRY_BATCH", "200"))
	cfg.RSSD.RetryBase = MustDuration(Getenv("RETRY_BASE", "2s"))
	cfg.RSSD.RetryMax = MustDuration(Getenv("RETRY_MAX", "2m"))
	cfg.RSSD.RetryMaxAttempts = MustInt(Getenv("RETRY_MAX_ATTEMPTS", "8"))

	return cfg
}

func Getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func MustInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Panicf("invalid int value: %q", s)
	}
	return v
}

func MustFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Panicf("invalid float value: %q", s)
	}
	return v
}

func MustDuration(s string) time.Duration {
	v, err := time.ParseDuration(s)
	if err != nil {
		log.Panicf("invalid duration value: %q", s)
	}
	return v
}

func MustBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Panicf("invalid bool value: %q", s)
	}
	return v
}

func (c EmailConfig) SMTPEnabled() bool {
	return c.Address != "" && c.SMTPHost != "" && c.SMTPUsername != "" && c.SMTPPassword != ""
}

func (c EmailConfig) IMAPEnabled() bool {
	return c.IMAPHost != "" && c.IMAPUsername != "" && c.IMAPPassword != ""
}
