package main

import (
	"context"
	"os"

	"github.com/111zxc/rss-communicator/internal/app"
	"github.com/111zxc/rss-communicator/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := app.NewAppLogger("tg-bot")

	ctx := app.SignalContext()
	if err := app.RunTGBot(ctx, cfg, log); err != nil && err != context.Canceled {
		log.Error("tg-bot fatal", "err", err)
		os.Exit(1)
	}
}
