package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type ContactSender interface {
	Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error
}

type ContactDeliveryService struct {
	contacts repository.ContactsRepository
	sender   ContactSender
}

func NewContactDeliveryService(contacts repository.ContactsRepository, sender ContactSender) *ContactDeliveryService {
	return &ContactDeliveryService{contacts: contacts, sender: sender}
}

type TestSendInput struct {
	FeedName string
	FeedURL  string
	Title    string
	Link     string
	Summary  *string
	Author   *string
}

func (s *ContactDeliveryService) TestSend(ctx context.Context, contactID string, in TestSendInput) error {
	if contactID == "" {
		return ErrBadRequest
	}

	contact, err := s.contacts.GetByID(ctx, contactID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	feedName := in.FeedName
	if feedName == "" {
		feedName = "Test feed"
	}
	feedURL := in.FeedURL
	if feedURL == "" {
		feedURL = "https://example.com/feed"
	}
	title := in.Title
	if title == "" {
		title = "Test item"
	}
	link := in.Link
	if link == "" {
		link = "https://example.com/items/test"
	}

	feed := domain.Feed{
		ID:   "test-feed",
		Name: feedName,
		URL:  feedURL,
	}
	items := []domain.Item{{
		ID:      "test-item",
		FeedID:  feed.ID,
		Title:   title,
		Link:    link,
		Summary: in.Summary,
		Author:  in.Author,
	}}

	return s.sender.Send(ctx, contact, feed, items)
}
