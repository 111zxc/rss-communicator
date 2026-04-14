package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type Scheduler struct {
	db     repository.Store
	outbox repository.OutboxRepository
	log    *slog.Logger
	tick   time.Duration
	limit  int
}

func NewScheduler(db repository.Store, outbox repository.OutboxRepository, log *slog.Logger, tick time.Duration, limit int) *Scheduler {
	return &Scheduler{db: db, outbox: outbox, log: log, tick: tick, limit: limit}
}

func (s *Scheduler) Run(ctx context.Context) error {
	t := time.NewTicker(s.tick)
	defer t.Stop()

	s.log.Info("scheduler started", "tick", s.tick)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			now := time.Now().UTC()
			feeds, err := s.db.Feeds().ListDue(ctx, now, s.limit)
			if err != nil {
				s.log.Error("scheduler list due feeds failed", "err", err)
				continue
			}
			for _, f := range feeds {
				_ = s.outbox.Enqueue(ctx, string(queue.TopicFetch), queue.FetchJob{FeedID: f.ID}, now)
			}
		}
	}
}
