package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/111zxc/rss-communicator/internal/app"
	"github.com/111zxc/rss-communicator/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := config.NewLogger(config.LogConfig{
		Level:   config.Getenv("LOG_LEVEL", "info"),
		Format:  config.Getenv("LOG_FORMAT", "text"),
		AddSrc:  config.Getenv("LOG_ADD_SOURCE", "") == "true",
		Service: "tg-bot",
	})

	ctx := signalContext()
	if err := app.RunTGBot(ctx, cfg, log); err != nil && err != context.Canceled {
		log.Error("tg-bot fatal", "err", err)
		os.Exit(1)
	}
}

func signalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
