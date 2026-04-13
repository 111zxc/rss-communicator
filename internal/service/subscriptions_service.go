package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type SubscriptionService struct {
	subs     repository.SubscriptionsRepository
	feeds    repository.FeedsRepository
	contacts repository.ContactsRepository
}

func NewSubscriptionService(subs repository.SubscriptionsRepository, feeds repository.FeedsRepository, contacts repository.ContactsRepository) *SubscriptionService {
	return &SubscriptionService{subs: subs, feeds: feeds, contacts: contacts}
}

func (s *SubscriptionService) ListByFeed(ctx context.Context, feedID string) ([]domain.Subscription, error) {
	if strings.TrimSpace(feedID) == "" {
		return nil, ErrBadRequest
	}
	if _, err := s.feeds.GetByID(ctx, feedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.subs.ListByFeed(ctx, feedID)
}

func (s *SubscriptionService) ListByContact(ctx context.Context, contactID string) ([]domain.Subscription, error) {
	if strings.TrimSpace(contactID) == "" {
		return nil, ErrBadRequest
	}
	if _, err := s.contacts.GetByID(ctx, contactID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.subs.ListByContact(ctx, contactID)
}

func (s *SubscriptionService) Bind(ctx context.Context, feedID, contactID string) error {
	if strings.TrimSpace(feedID) == "" || strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	if _, err := s.feeds.GetByID(ctx, feedID); err != nil {
		return err
	}
	if _, err := s.contacts.GetByID(ctx, contactID); err != nil {
		return err
	}
	return s.subs.Add(ctx, feedID, contactID)
}

func (s *SubscriptionService) Unbind(ctx context.Context, feedID, contactID string) error {
	if strings.TrimSpace(feedID) == "" || strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	return s.subs.Remove(ctx, feedID, contactID)
}
