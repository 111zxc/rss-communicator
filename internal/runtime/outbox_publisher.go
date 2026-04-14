package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type OutboxPublisher struct {
	repo      repository.OutboxRepository
	q         queue.Queue
	log       *slog.Logger
	tick      time.Duration
	lease     time.Duration
	limit     int
	retryBack Backoff
}

func NewOutboxPublisher(repo repository.OutboxRepository, q queue.Queue, log *slog.Logger, tick time.Duration, lease time.Duration, limit int, retryBack Backoff) *OutboxPublisher {
	if tick <= 0 {
		tick = time.Second
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	if limit <= 0 {
		limit = 100
	}
	if retryBack.Base <= 0 {
		retryBack.Base = time.Second
	}
	if retryBack.Max <= 0 {
		retryBack.Max = time.Minute
	}

	return &OutboxPublisher{
		repo:      repo,
		q:         q,
		log:       log,
		tick:      tick,
		lease:     lease,
		limit:     limit,
		retryBack: retryBack,
	}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	t := time.NewTicker(p.tick)
	defer t.Stop()

	p.log.Info("outbox publisher started", "tick", p.tick, "lease", p.lease, "limit", p.limit)

	for {
		if err := p.flushOnce(ctx); err != nil && err != context.Canceled {
			p.log.Error("outbox publisher flush failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (p *OutboxPublisher) flushOnce(ctx context.Context) error {
	now := time.Now().UTC()
	msgs, err := p.repo.ClaimBatch(ctx, now, now.Add(p.lease), p.limit)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if err := p.q.Publish(ctx, queue.Topic(msg.Topic), msg.Payload); err != nil {
			next := time.Now().UTC().Add(p.retryBack.Delay(msg.AttemptCount + 1))
			if markErr := p.repo.MarkFailed(ctx, msg.ID, err.Error(), next); markErr != nil {
				p.log.Error("outbox mark failed failed", "id", msg.ID, "err", markErr)
			}
			continue
		}
		if err := p.repo.MarkPublished(ctx, msg.ID); err != nil {
			p.log.Error("outbox mark published failed", "id", msg.ID, "err", err)
		}
	}

	return nil
}
