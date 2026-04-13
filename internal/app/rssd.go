package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/config"
	"github.com/111zxc/rss-communicator/internal/handler"
	"github.com/111zxc/rss-communicator/internal/repository/postgres"
	"github.com/111zxc/rss-communicator/internal/rss"
	"github.com/111zxc/rss-communicator/internal/runtime"
	"github.com/111zxc/rss-communicator/internal/runtime/queue/memory"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
	"github.com/111zxc/rss-communicator/internal/senders"
	httpsender "github.com/111zxc/rss-communicator/internal/senders/http"
	tgsender "github.com/111zxc/rss-communicator/internal/senders/telegram"
	"github.com/111zxc/rss-communicator/internal/service"
)

type RSSD struct {
	db  *postgres.DB
	log *slog.Logger
}

func NewRSSD(db *postgres.DB, log *slog.Logger) *RSSD {
	return &RSSD{db: db, log: log}
}

func RunRSSD(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	// DB
	db, err := postgres.New(cfg.DB.DSN)
	if err != nil {
		log.Error("db connect failed", "err", err)
		return err
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Error("db ping failed", "err", err)
		return err
	}

	// In-proc queue
	q := memory.New()
	defer func() { _ = q.Close() }()

	// Fetcher
	fetcher := rss.NewFetcher()

	// Fetch worker
	fw := worker.NewFetchWorker(db, q, log, fetcher)
	if err := fw.Subscribe(ctx); err != nil {
		log.Error("fetch worker subscribe failed", "err", err)
		return err
	}

	// Retry scheduler
	rs := runtime.NewRetryScheduler(
		db.Deliveries(),
		q,
		log,
		cfg.RSSD.RetryScheduleTick,
		cfg.RSSD.RetryBatch,
		cfg.RSSD.RetryMaxAttempts,
	)
	go func() {
		_ = rs.Run(ctx)
	}()

	var telegramDeliverySender senders.Sender
	if cfg.Telegram.BotToken != "" {
		tgAPI, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
		if err != nil {
			log.Error("telegram init failed", "err", err)
			return err
		}
		telegramDeliverySender = tgsender.New(tgAPI)
	} else {
		log.Warn("telegram sender is disabled: TELEGRAM_BOT_TOKEN is empty")
	}

	httpDeliverySender := httpsender.New(db.Contacts(), &http.Client{Timeout: cfg.RSSD.HTTPTimeout})
	sender := senders.NewRouter(telegramDeliverySender, httpDeliverySender)

	// Delivery worker pool + rate limit
	limiter := runtime.NewTokenBucket(cfg.RSSD.TelegramRPS, cfg.RSSD.TelegramBurst)
	dw := worker.NewDeliverWorker(
		db, q, log, sender, limiter,
		worker.DeliverWorkerConfig{
			Workers:     cfg.RSSD.DeliveryWorkers,
			RetryBase:   cfg.RSSD.RetryBase,
			RetryMax:    cfg.RSSD.RetryMax,
			MaxAttempts: cfg.RSSD.RetryMaxAttempts,
		},
	)
	if err := dw.Subscribe(ctx); err != nil {
		log.Error("deliver worker subscribe failed", "err", err)
		return err
	}

	// Scheduler (publishes fetch jobs)
	s := runtime.NewScheduler(db, q, log, 5*time.Second, 50)
	go func() {
		if err := s.Run(ctx); err != nil && err != context.Canceled {
			log.Error("scheduler stopped with error", "err", err)
		}
	}()

	log.Info("rssd started",
		"schedule_tick", "5s",
		"fetch_batch_limit", 50,
		"deliver_workers", 4,
		"tg_rps", 5.0,
		"tg_burst", 10,
		"http_timeout", cfg.RSSD.HTTPTimeout.String(),
	)

	// --- services ---
	clock := service.SystemClock{}
	feedSvc := service.NewFeedService(db.Feeds(), clock)
	contactSvc := service.NewContactService(db.Contacts())
	contactDeliverySvc := service.NewContactDeliveryService(db.Contacts(), sender)
	subSvc := service.NewSubscriptionService(db.Subscriptions(), db.Feeds(), db.Contacts())

	// --- handlers ---
	h := handler.New(feedSvc, contactSvc, contactDeliverySvc, subSvc)

	// --- router ---
	router := NewRouter(h)

	// --- http server ---
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		log.Info("http server started", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server failed", "err", err)
		}
	}()

	// graceful shutdown
	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxShutdown)
	}()

	<-ctx.Done()
	log.Info("rssd stopped")
	return ctx.Err()
}
