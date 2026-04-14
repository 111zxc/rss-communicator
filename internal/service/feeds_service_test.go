package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestFeedServiceCreateValidatesInput(t *testing.T) {
	svc := NewFeedService(&feedRepoStub{}, fixedClock{})

	tests := []struct {
		name  string
		input CreateFeedInput
	}{
		{
			name: "missing name",
			input: CreateFeedInput{
				URL:             "https://example.com/rss",
				IntervalSeconds: 60,
			},
		},
		{
			name: "missing url",
			input: CreateFeedInput{
				Name:            "Example",
				IntervalSeconds: 60,
			},
		},
		{
			name: "invalid url",
			input: CreateFeedInput{
				Name:            "Example",
				URL:             "://bad",
				IntervalSeconds: 60,
			},
		},
		{
			name: "bad interval",
			input: CreateFeedInput{
				Name:            "Example",
				URL:             "https://example.com/rss",
				IntervalSeconds: 0,
			},
		},
		{
			name: "bad batch window",
			input: CreateFeedInput{
				Name:            "Example",
				URL:             "https://example.com/rss",
				IntervalSeconds: 60,
				BatchEnabled:    true,
				BatchWindowSecs: 59,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.input)
			if !errors.Is(err, ErrBadRequest) {
				t.Fatalf("expected ErrBadRequest, got %v", err)
			}
		})
	}
}

func TestFeedServiceCreateDefaultsBatchWindow(t *testing.T) {
	repo := &feedRepoStub{}
	svc := NewFeedService(repo, fixedClock{})

	got, err := svc.Create(context.Background(), CreateFeedInput{
		Name:            "Example",
		URL:             "https://example.com/rss",
		IntervalSeconds: 300,
		Enabled:         true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got.BatchWindowSecs != 3600 {
		t.Fatalf("expected default batch window 3600, got %d", got.BatchWindowSecs)
	}
	if repo.created.BatchWindowSecs != 3600 {
		t.Fatalf("expected repo to receive default batch window 3600, got %d", repo.created.BatchWindowSecs)
	}
}

func TestFeedServiceUpdateBatching(t *testing.T) {
	repo := &feedRepoStub{
		getByIDFeed: domain.Feed{
			ID:              "feed-1",
			Name:            "Example",
			URL:             "https://example.com/rss",
			IntervalSeconds: 300,
			BatchEnabled:    false,
			BatchWindowSecs: 3600,
		},
	}
	svc := NewFeedService(repo, fixedClock{})
	enabled := true
	window := 1800

	got, err := svc.Update(context.Background(), "feed-1", UpdateFeedInput{
		BatchEnabled:    &enabled,
		BatchWindowSecs: &window,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected UpdateBatching to be called")
	}
	if !got.BatchEnabled || got.BatchWindowSecs != 1800 {
		t.Fatalf("unexpected updated feed: %+v", got)
	}
}

func TestFeedServiceDeleteRejectsBlankID(t *testing.T) {
	svc := NewFeedService(&feedRepoStub{}, fixedClock{})

	err := svc.Delete(context.Background(), " ")
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

type fixedClock struct{}

func (fixedClock) NowUTC() time.Time {
	return time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
}

type feedRepoStub struct {
	created      domain.Feed
	getByIDFeed  domain.Feed
	updateCalled bool
}

func (r *feedRepoStub) Create(_ context.Context, f domain.Feed) (domain.Feed, error) {
	r.created = f
	f.ID = "feed-1"
	return f, nil
}

func (r *feedRepoStub) ListDue(context.Context, time.Time, int) ([]domain.Feed, error) {
	return nil, nil
}

func (r *feedRepoStub) MarkFetched(context.Context, string, time.Time, time.Time, *string, *string) error {
	return nil
}

func (r *feedRepoStub) MarkFetchError(context.Context, string, string) error {
	return nil
}

func (r *feedRepoStub) MarkInitialized(context.Context, string, time.Time) error {
	return nil
}

func (r *feedRepoStub) GetByID(context.Context, string) (domain.Feed, error) {
	return r.getByIDFeed, nil
}

func (r *feedRepoStub) UpdateBatching(_ context.Context, feedID string, batchEnabled bool, batchWindowSecs int) (domain.Feed, error) {
	r.updateCalled = true
	r.getByIDFeed.ID = feedID
	r.getByIDFeed.BatchEnabled = batchEnabled
	r.getByIDFeed.BatchWindowSecs = batchWindowSecs
	return r.getByIDFeed, nil
}

func (r *feedRepoStub) List(context.Context, int, int) ([]domain.Feed, int, error) {
	return nil, 0, nil
}

func (r *feedRepoStub) Delete(context.Context, string) error {
	return nil
}
