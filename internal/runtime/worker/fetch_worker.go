package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository/postgres"
	"github.com/111zxc/rss-communicator/internal/rss"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type FetchWorker struct {
	db      *postgres.DB
	q       queue.Queue
	log     *slog.Logger
	fetcher *rss.Fetcher
}

func NewFetchWorker(db *postgres.DB, q queue.Queue, log *slog.Logger, fetcher *rss.Fetcher) *FetchWorker {
	return &FetchWorker{db: db, q: q, log: log, fetcher: fetcher}
}

func (w *FetchWorker) Subscribe(ctx context.Context) error {
	return w.q.Subscribe(ctx, queue.TopicFetch, func(hctx context.Context, m queue.Message) error {
		var job queue.FetchJob
		if err := json.Unmarshal(m.Data, &job); err != nil {
			w.log.Error("bad fetch job", "err", err)
			return nil
		}
		w.process(hctx, job.FeedID)
		return nil
	})
}

func (w *FetchWorker) process(ctx context.Context, feedID string) {
	f, err := w.db.Feeds().GetByID(ctx, feedID)
	if err != nil {
		w.log.Error("feed get failed", "feed_id", feedID, "err", err)
		return
	}
	if !f.Enabled {
		return
	}

	res, err := w.fetcher.Fetch(ctx, f.URL, f.ETag, f.LastModified)
	if err != nil {
		_ = w.db.Feeds().MarkFetchError(ctx, f.ID, err.Error())
		w.log.Warn("fetch failed", "feed", f.Name, "err", err)
		return
	}

	nextAt := time.Now().UTC().Add(time.Duration(f.IntervalSeconds) * time.Second)
	_ = w.db.Feeds().MarkFetched(ctx, f.ID, time.Now().UTC(), nextAt, res.ETag, res.LastModified)

	if res.NotModified {
		return
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		w.log.Warn("bad status", "status", res.StatusCode, "feed", f.Name)
		return
	}

	parsed, err := rss.Parse(res.Body)
	if err != nil {
		w.log.Warn("parse failed", "feed", f.Name, "err", err)
		return
	}

	items := make([]domain.Item, 0, len(parsed))
	for _, p := range parsed {
		items = append(items, domain.Item{
			FeedID:      f.ID,
			ExternalID:  p.ExternalID,
			UniqKey:     p.UniqKey,
			Title:       p.Title,
			Link:        p.Link,
			Summary:     p.Summary,
			Author:      p.Author,
			PublishedAt: p.PublishedAt,
		})
	}

	inserted, err := w.db.Items().InsertMany(ctx, items)
	if err != nil {
		w.log.Error("insert items failed", "feed", f.Name, "err", err)
		return
	}

	if f.InitializedAt == nil { // warm start
		_ = w.db.Feeds().MarkInitialized(ctx, f.ID, time.Now().UTC())
		w.log.Info("warm start: initialized", "feed", f.Name, "inserted", len(inserted))
		return
	}

	if len(inserted) == 0 {
		return
	}

	subs, err := w.db.Subscriptions().ListByFeed(ctx, f.ID)
	if err != nil {
		w.log.Error("list subs failed", "feed", f.Name, "err", err)
		return
	}

	for _, it := range inserted {
		for _, s := range subs {
			availableAt := time.Now().UTC()
			if f.BatchEnabled {
				availableAt = nextBatchWindow(time.Now().UTC(), f.BatchWindowSecs)
			}

			created, deliveryID, err := w.db.Deliveries().CreatePendingIfNotExists(ctx, s.ContactID, it.ID, availableAt)
			if err != nil || !created {
				continue
			}
			if f.BatchEnabled {
				continue
			}
			b, _ := json.Marshal(queue.DeliverJob{DeliveryID: deliveryID})
			_ = w.q.Publish(ctx, queue.TopicDeliver, b)
		}
	}
}

func nextBatchWindow(now time.Time, windowSecs int) time.Time {
	if windowSecs < 60 {
		windowSecs = 3600
	}
	window := time.Duration(windowSecs) * time.Second
	return now.UTC().Truncate(window).Add(window)
}
