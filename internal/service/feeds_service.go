package service

import (
	"context"
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
	if _, err := url.ParseRequestURI(in.URL); err != nil {
		return domain.Feed{}, ErrBadRequest
	}

	f := domain.Feed{
		Name:            in.Name,
		URL:             in.URL,
		IntervalSeconds: in.IntervalSeconds,
		Enabled:         in.Enabled,
		CreatedAt:       s.now.NowUTC(),
		UpdatedAt:       s.now.NowUTC(),
	}

	return s.feeds.Create(ctx, f)
}

func (s *FeedService) Delete(ctx context.Context, feedID string) error {
	if strings.TrimSpace(feedID) == "" {
		return ErrBadRequest
	}
	return s.feeds.Delete(ctx, feedID)
}

type SystemClock struct{}

func (SystemClock) NowUTC() time.Time { return time.Now().UTC() }
