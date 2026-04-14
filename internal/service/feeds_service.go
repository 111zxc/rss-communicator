package service

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

var (
	ErrNotFound   = errors.New("not_found")
	ErrBadRequest = errors.New("bad_request")
)

type FeedService struct {
	feeds repository.FeedsRepository
	now   repository.Clock
}

func NewFeedService(feeds repository.FeedsRepository, now repository.Clock) *FeedService {
	return &FeedService{feeds: feeds, now: now}
}

type CreateFeedInput struct {
	Name            string
	URL             string
	IntervalSeconds int
	Enabled         bool
	BatchEnabled    bool
	BatchWindowSecs int
}

type UpdateFeedInput struct {
	BatchEnabled    *bool
	BatchWindowSecs *int
}

func (s *FeedService) List(ctx context.Context, limit, offset int) ([]domain.Feed, int, error) {
	return s.feeds.List(ctx, limit, offset)
}

func (s *FeedService) Create(ctx context.Context, in CreateFeedInput) (domain.Feed, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)

	if in.Name == "" || in.URL == "" {
		return domain.Feed{}, ErrBadRequest
	}
	if in.IntervalSeconds <= 0 {
		return domain.Feed{}, ErrBadRequest
	}
	if in.BatchEnabled && in.BatchWindowSecs < 60 {
		return domain.Feed{}, ErrBadRequest
	}
	if !in.BatchEnabled && in.BatchWindowSecs == 0 {
		in.BatchWindowSecs = 3600
	}
	if in.BatchWindowSecs < 60 {
		in.BatchWindowSecs = 3600
	}
	if _, err := url.ParseRequestURI(in.URL); err != nil {
		return domain.Feed{}, ErrBadRequest
	}

	f := domain.Feed{
		Name:            in.Name,
		URL:             in.URL,
		IntervalSeconds: in.IntervalSeconds,
		Enabled:         in.Enabled,
		BatchEnabled:    in.BatchEnabled,
		BatchWindowSecs: in.BatchWindowSecs,
		CreatedAt:       s.now.NowUTC(),
		UpdatedAt:       s.now.NowUTC(),
	}

	return s.feeds.Create(ctx, f)
}

func (s *FeedService) Delete(ctx context.Context, feedID string) error {
	if strings.TrimSpace(feedID) == "" {
		return ErrBadRequest
	}
	err := s.feeds.Delete(ctx, feedID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *FeedService) Update(ctx context.Context, feedID string, in UpdateFeedInput) (domain.Feed, error) {
	if strings.TrimSpace(feedID) == "" {
		return domain.Feed{}, ErrBadRequest
	}

	f, err := s.feeds.GetByID(ctx, feedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Feed{}, ErrNotFound
		}
		return domain.Feed{}, err
	}

	if in.BatchEnabled != nil {
		f.BatchEnabled = *in.BatchEnabled
	}
	if in.BatchWindowSecs != nil {
		f.BatchWindowSecs = *in.BatchWindowSecs
	}
	if f.BatchWindowSecs < 60 {
		return domain.Feed{}, ErrBadRequest
	}

	updated, err := s.feeds.UpdateBatching(ctx, feedID, f.BatchEnabled, f.BatchWindowSecs)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, ErrNotFound
	}
	return updated, err
}

type SystemClock struct{}

func (SystemClock) NowUTC() time.Time { return time.Now().UTC() }
