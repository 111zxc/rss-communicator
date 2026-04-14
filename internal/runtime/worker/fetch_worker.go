package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/rss"
	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type FetchWorker struct {
	db      repository.Database
	q       queue.Queue
	log     *slog.Logger
	fetcher *rss.Fetcher
}

func NewFetchWorker(db repository.Database, q queue.Queue, _ repository.OutboxRepository, log *slog.Logger, fetcher *rss.Fetcher) *FetchWorker {
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

	now := time.Now().UTC()
	nextAt := now.Add(time.Duration(f.IntervalSeconds) * time.Second)
	if res.NotModified {
		_ = w.db.Feeds().MarkFetched(ctx, f.ID, now, nextAt, res.ETag, res.LastModified)
		return
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		_ = w.db.Feeds().MarkFetched(ctx, f.ID, now, nextAt, res.ETag, res.LastModified)
		w.log.Warn("bad status", "status", res.StatusCode, "feed", f.Name)
		return
	}

	parsed, err := rss.Parse(res.Body)
	if err != nil {
		_ = w.db.Feeds().MarkFetched(ctx, f.ID, now, nextAt, res.ETag, res.LastModified)
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

	var insertedCount int
	err = w.db.WithinTx(ctx, func(store repository.Store) error {
		if err := store.Feeds().MarkFetched(ctx, f.ID, now, nextAt, res.ETag, res.LastModified); err != nil {
			return err
		}

		inserted, err := store.Items().InsertMany(ctx, items)
		if err != nil {
			return err
		}
		insertedCount = len(inserted)

		if f.InitializedAt == nil {
			return store.Feeds().MarkInitialized(ctx, f.ID, now)
		}
		if len(inserted) == 0 {
			return nil
		}

		subs, err := store.Subscriptions().ListByFeed(ctx, f.ID)
		if err != nil {
			return err
		}

		for _, it := range inserted {
			for _, s := range subs {
				availableAt := now
				if f.BatchEnabled {
					availableAt = nextBatchWindow(now, f.BatchWindowSecs)
				}

				created, deliveryID, err := store.Deliveries().CreatePendingIfNotExists(ctx, s.ContactID, it.ID, availableAt)
				if err != nil {
					return err
				}
				if !created || f.BatchEnabled {
					continue
				}
				if err := store.Outbox().Enqueue(ctx, string(queue.TopicDeliver), queue.DeliverJob{DeliveryID: deliveryID}, availableAt); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = w.db.Feeds().MarkFetchError(ctx, f.ID, err.Error())
		w.log.Error("persist fetched items failed", "feed", f.Name, "err", err)
		return
	}

	if f.InitializedAt == nil {
		w.log.Info("warm start: initialized", "feed", f.Name, "inserted", insertedCount)
	}
}

func nextBatchWindow(now time.Time, windowSecs int) time.Time {
	if windowSecs < 60 {
		windowSecs = 3600
	}
	window := time.Duration(windowSecs) * time.Second
	return now.UTC().Truncate(window).Add(window)
}
