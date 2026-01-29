package service

import (
	"context"
	"strings"

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
