package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/111zxc/rss-communicator/internal/config"
)

func NewAppLogger(appName string) *slog.Logger {
	return config.NewLogger(config.LogConfig{
		Level:   config.Getenv("LOG_LEVEL", "info"),
		Format:  config.Getenv("LOG_FORMAT", "text"),
		AddSrc:  config.Getenv("LOG_ADD_SOURCE", "") == "true",
		Service: appName,
	})
}

func SignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
