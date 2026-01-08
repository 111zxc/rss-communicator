package config

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type LogConfig struct {
	Level   string // debug|info|warn|error
	Format  string // text|json
	AddSrc  bool   // add source file:line
	Service string // rssd / tg-bot
}

func NewLogger(cfg LogConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSrc,
	}

	var w io.Writer = os.Stdout
	var h slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	default:
		h = slog.NewTextHandler(w, opts)
	}

	l := slog.New(h)
	if cfg.Service != "" {
		l = l.With("service", cfg.Service)
	}
	return l
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
