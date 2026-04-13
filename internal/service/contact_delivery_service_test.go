package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestContactDeliveryServiceTestSendUsesDefaults(t *testing.T) {
	sender := &contactDeliverySenderStub{}
	contacts := &contactDeliveryContactsStub{
		contact: domain.Contact{ID: "contact-1", Type: domain.ContactHTTP, Status: domain.ContactActive},
	}
	svc := NewContactDeliveryService(contacts, sender)

	err := svc.TestSend(context.Background(), "contact-1", TestSendInput{})
	if err != nil {
		t.Fatalf("TestSend returned error: %v", err)
	}
	if sender.feed.Name != "Test feed" || len(sender.items) != 1 || sender.items[0].Title != "Test item" {
		t.Fatalf("unexpected test payload: feed=%+v items=%+v", sender.feed, sender.items)
	}
}

func TestContactDeliveryServiceTestSendMapsMissingContact(t *testing.T) {
	svc := NewContactDeliveryService(&contactDeliveryContactsStub{err: sql.ErrNoRows}, &contactDeliverySenderStub{})

	err := svc.TestSend(context.Background(), "missing", TestSendInput{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

type contactDeliverySenderStub struct {
	contact domain.Contact
	feed    domain.Feed
	items   []domain.Item
}

func (s *contactDeliverySenderStub) Send(_ context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	s.contact = c
	s.feed = feed
	s.items = items
	return nil
}

type contactDeliveryContactsStub struct {
	contact domain.Contact
	err     error
}

func (s *contactDeliveryContactsStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) CreateTelegram(context.Context, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) UpdateTelegram(context.Context, string, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) CreateHTTP(context.Context, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) UpdateHTTP(context.Context, string, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}

func (s *contactDeliveryContactsStub) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactDeliveryContactsStub) GetByID(context.Context, string) (domain.Contact, error) {
	if s.err != nil {
		return domain.Contact{}, s.err
	}
	return s.contact, nil
}

func (s *contactDeliveryContactsStub) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return nil, 0, nil
}

func (s *contactDeliveryContactsStub) Delete(context.Context, string) error {
	return nil
}
