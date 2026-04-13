package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/config"
	appemail "github.com/111zxc/rss-communicator/internal/email"
	"github.com/111zxc/rss-communicator/internal/handler"
	"github.com/111zxc/rss-communicator/internal/repository/postgres"
	"github.com/111zxc/rss-communicator/internal/rss"
	"github.com/111zxc/rss-communicator/internal/runtime"
	"github.com/111zxc/rss-communicator/internal/runtime/queue/memory"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
	"github.com/111zxc/rss-communicator/internal/senders"
	emailsender "github.com/111zxc/rss-communicator/internal/senders/email"
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

	var emailDeliverySender senders.Sender
	if cfg.Email.SMTPEnabled() {
		emailDeliverySender = emailsender.New(db.Contacts(), cfg.Email)
	} else {
		log.Warn("email sender is disabled: smtp settings are incomplete")
	}

	httpDeliverySender := httpsender.New(db.Contacts(), &http.Client{Timeout: cfg.RSSD.HTTPTimeout})
	sender := senders.NewRouter(telegramDeliverySender, emailDeliverySender, httpDeliverySender)

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
	s := runtime.NewScheduler(db, q, log, cfg.RSSD.ScheduleTick, cfg.RSSD.FetchBatch)
	go func() {
		if err := s.Run(ctx); err != nil && err != context.Canceled {
			log.Error("scheduler stopped with error", "err", err)
		}
	}()

	log.Info("rssd started",
		"schedule_tick", cfg.RSSD.ScheduleTick.String(),
		"fetch_batch_limit", cfg.RSSD.FetchBatch,
		"deliver_workers", cfg.RSSD.DeliveryWorkers,
		"tg_rps", cfg.RSSD.TelegramRPS,
		"tg_burst", cfg.RSSD.TelegramBurst,
		"http_timeout", cfg.RSSD.HTTPTimeout.String(),
	)

	// --- services ---
	clock := service.SystemClock{}
	feedSvc := service.NewFeedService(db.Feeds(), clock)
	contactSvc := service.NewContactService(db.Contacts())
	contactDeliverySvc := service.NewContactDeliveryService(db.Contacts(), sender)
	subSvc := service.NewSubscriptionService(db.Subscriptions(), db.Feeds(), db.Contacts())
	groupSvc := service.NewGroupService(db.Groups(), db.Feeds(), db.Contacts(), db.Subscriptions())
	regCodeSvc := service.NewRegistrationCodeService(db.RegistrationCodes(), db.Groups())
	regSvc := service.NewRegistrationService(db.Contacts(), db.RegistrationCodes(), db.Groups(), db.Subscriptions())

	if cfg.Email.IMAPEnabled() {
		mailbox := appemail.NewIMAPMailbox(appemail.IMAPConfig{
			Host:     cfg.Email.IMAPHost,
			Port:     cfg.Email.IMAPPort,
			Username: cfg.Email.IMAPUsername,
			Password: cfg.Email.IMAPPassword,
			Mailbox:  cfg.Email.IMAPMailbox,
		})
		poller := appemail.NewPoller(mailbox, regSvc, log, cfg.Email.IMAPPollInterval)
		go func() {
			if err := poller.Run(ctx); err != nil && err != context.Canceled {
				log.Error("email poller stopped with error", "err", err)
			}
		}()
	} else {
		log.Warn("email inbox registration is disabled: imap settings are incomplete")
	}

	// --- handlers ---
	h := handler.New(feedSvc, contactSvc, contactDeliverySvc, subSvc, groupSvc, regCodeSvc)

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
