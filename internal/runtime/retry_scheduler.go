package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type RetryLister interface {
	ListRetryDue(ctx context.Context, now time.Time, limit int, maxAttempts int) ([]string, error)
}

type RetryScheduler struct {
	lister      RetryLister
	outbox      repository.OutboxRepository
	log         *slog.Logger
	tick        time.Duration
	limit       int
	maxAttempts int
}

func NewRetryScheduler(lister RetryLister, outbox repository.OutboxRepository, log *slog.Logger, tick time.Duration, limit int, maxAttempts int) *RetryScheduler {
	return &RetryScheduler{
		lister:      lister,
		outbox:      outbox,
		log:         log,
		tick:        tick,
		limit:       limit,
		maxAttempts: maxAttempts,
	}
}

func (s *RetryScheduler) Run(ctx context.Context) error {
	t := time.NewTicker(s.tick)
	defer t.Stop()

	s.log.Info("retry scheduler started", "tick", s.tick, "limit", s.limit)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			now := time.Now().UTC()
			ids, err := s.lister.ListRetryDue(ctx, now, s.limit, s.maxAttempts)
			if err != nil {
				s.log.Error("list retry due failed", "err", err)
				continue
			}
			for _, id := range ids {
				_ = s.outbox.Enqueue(ctx, string(queue.TopicDeliver), queue.DeliverJob{DeliveryID: id}, now)
			}
		}
	}
}
