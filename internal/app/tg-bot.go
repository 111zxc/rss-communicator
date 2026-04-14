package app

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/config"
	"github.com/111zxc/rss-communicator/internal/database"
	"github.com/111zxc/rss-communicator/internal/service"
	"github.com/111zxc/rss-communicator/internal/telegram"
)

func RunTGBot(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	if cfg.Telegram.BotToken == "" {
		log.Error("TELEGRAM_BOT_TOKEN is empty")
		return context.Canceled
	}

	db, err := database.Open(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		log.Error("db connect failed", "err", err)
		return err
	}
	defer db.Close()

	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.Ping(pctx); err != nil {
		log.Error("db ping failed", "err", err)
		return err
	}

	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		log.Error("telegram init failed", "err", err)
		return err
	}
	api.Debug = false

	regSvc := service.NewRegistrationService(db.Contacts(), db.RegistrationCodes(), db.Groups(), db.Subscriptions())
	b := telegram.New(api, db, regSvc, log)

	log.Info("tg-bot started")
	if err := b.Run(ctx); err != nil && err != context.Canceled {
		log.Error("tg-bot exited with error", "err", err)
		return err
	}
	log.Info("tg-bot stopped")
	return ctx.Err()
}
