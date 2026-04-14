package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/runtime"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type DeliverySender interface {
	Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error
}

type DeliverWorkerConfig struct {
	Workers     int
	RetryBase   time.Duration
	RetryMax    time.Duration
	MaxAttempts int
}

type DeliverWorker struct {
	db          repository.Store
	q           queue.Queue
	log         *slog.Logger
	sender      DeliverySender
	limiter     *runtime.TokenBucket
	backoff     runtime.Backoff
	maxAttempts int
	sem         chan struct{}
}

func NewDeliverWorker(
	db repository.Store,
	q queue.Queue,
	log *slog.Logger,
	sender DeliverySender,
	limiter *runtime.TokenBucket,
	cfg DeliverWorkerConfig,
) *DeliverWorker {
	workers := cfg.Workers
	if workers < 1 {
		workers = 1
	}

	base := cfg.RetryBase
	if base <= 0 {
		base = 2 * time.Second
	}
	max := cfg.RetryMax
	if max <= 0 {
		max = 2 * time.Minute
	}
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 8
	}

	return &DeliverWorker{
		db:          db,
		q:           q,
		log:         log,
		sender:      sender,
		limiter:     limiter,
		backoff:     runtime.Backoff{Base: base, Max: max},
		maxAttempts: maxAttempts,
		sem:         make(chan struct{}, workers),
	}
}

func (w *DeliverWorker) Subscribe(ctx context.Context) error {
	return w.q.Subscribe(ctx, queue.TopicDeliver, func(hctx context.Context, m queue.Message) error {
		var job queue.DeliverJob
		if err := json.Unmarshal(m.Data, &job); err != nil {
			w.log.Error("bad deliver job", "err", err)
			return nil
		}

		select {
		case w.sem <- struct{}{}:
			go func() {
				defer func() { <-w.sem }()
				w.process(hctx, job.DeliveryID)
			}()
		case <-hctx.Done():
			return hctx.Err()
		}
		return nil
	})
}

func (w *DeliverWorker) process(ctx context.Context, deliveryID string) {
	contact, feed, batch, err := w.db.Deliveries().ClaimBatch(ctx, deliveryID, time.Now().UTC())
	if err != nil {
		w.log.Error("delivery claim failed", "delivery_id", deliveryID, "err", err)
		return
	}
	if len(batch) == 0 {
		return
	}

	if contact.Status != domain.ContactActive {
		ids := deliveryIDs(batch)
		_ = w.db.Deliveries().MarkManyFailed(ctx, ids, "contact is not active", nil)
		return
	}

	items := make([]domain.Item, 0, len(batch))
	attemptCount := 0
	for _, entry := range batch {
		items = append(items, entry.Item)
		if entry.Delivery.AttemptCount > attemptCount {
			attemptCount = entry.Delivery.AttemptCount
		}
	}

	if err := w.limiter.Wait(ctx); err != nil {
		return
	}

	err = w.sender.Send(ctx, contact, feed, items)
	ids := deliveryIDs(batch)
	if err == nil {
		_ = w.db.Deliveries().MarkManySent(ctx, ids, time.Now().UTC())
		return
	}

	next := time.Now().UTC().Add(w.backoff.Delay(attemptCount + 1))
	if isPermanent(err) || attemptCount >= w.maxAttempts {
		_ = w.db.Deliveries().MarkManyFailed(ctx, ids, err.Error(), nil)
		w.log.Warn("delivery dead/failed", "delivery_id", deliveryID, "err", err)
		return
	}

	_ = w.db.Deliveries().MarkManyFailed(ctx, ids, err.Error(), &next)
}

func deliveryIDs(batch []domain.DeliveryWithItem) []string {
	out := make([]string, 0, len(batch))
	for _, entry := range batch {
		out = append(out, entry.Delivery.ID)
	}
	return out
}

func isPermanent(err error) bool {
	if err == nil {
		return false
	}
	var e *PermanentError
	if errors.As(err, &e) {
		return true
	}
	return false
}

type PermanentError struct{ Msg string }

func (e *PermanentError) Error() string { return e.Msg }
