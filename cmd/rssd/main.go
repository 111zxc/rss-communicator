package main

import (
	"context"
	"os"

	"github.com/111zxc/rss-communicator/internal/app"
	"github.com/111zxc/rss-communicator/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := app.NewAppLogger("rssd")

	ctx := app.SignalContext()
	if err := app.RunRSSD(ctx, cfg, log); err != nil && err != context.Canceled {
		log.Error("rssd fatal", "err", err)
		os.Exit(1)
	}
}
