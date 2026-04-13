package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestSubscriptionServiceBindValidatesIDs(t *testing.T) {
	svc := NewSubscriptionService(&subsRepoStub{}, &feedLookupStub{}, &contactsRepoStub{})

	err := svc.Bind(context.Background(), "", "contact-1")
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestSubscriptionServiceListByContact(t *testing.T) {
	subs := &subsRepoStub{
		listByContact: []domain.Subscription{{FeedID: "feed-1", ContactID: "contact-1"}},
	}
	contacts := &contactsRepoStub{contact: domain.Contact{ID: "contact-1", Type: domain.ContactHTTP}}
	svc := NewSubscriptionService(subs, &feedLookupStub{}, contacts)

	got, err := svc.ListByContact(context.Background(), "contact-1")
	if err != nil {
		t.Fatalf("ListByContact returned error: %v", err)
	}
	if len(got) != 1 || got[0].FeedID != "feed-1" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
}

func TestSubscriptionServiceListByContactMapsMissingContact(t *testing.T) {
	svc := NewSubscriptionService(&subsRepoStub{}, &feedLookupStub{}, &contactsRepoStub{err: sql.ErrNoRows})

	_, err := svc.ListByContact(context.Background(), "contact-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubscriptionServiceListByFeed(t *testing.T) {
	subs := &subsRepoStub{
		listByFeed: []domain.Subscription{{FeedID: "feed-1", ContactID: "contact-1"}},
	}
	feeds := &feedLookupStub{feed: domain.Feed{ID: "feed-1"}}
	svc := NewSubscriptionService(subs, feeds, &contactsRepoStub{})

	got, err := svc.ListByFeed(context.Background(), "feed-1")
	if err != nil {
		t.Fatalf("ListByFeed returned error: %v", err)
	}
	if len(got) != 1 || got[0].ContactID != "contact-1" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
}

func TestSubscriptionServiceListByFeedMapsMissingFeed(t *testing.T) {
	svc := NewSubscriptionService(&subsRepoStub{}, &feedLookupStub{err: sql.ErrNoRows}, &contactsRepoStub{})

	_, err := svc.ListByFeed(context.Background(), "feed-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubscriptionServiceBindChecksDependenciesAndAddsSubscription(t *testing.T) {
	subs := &subsRepoStub{}
	feeds := &feedLookupStub{feed: domain.Feed{ID: "feed-1"}}
	contacts := &contactsRepoStub{contact: domain.Contact{ID: "contact-1"}}
	svc := NewSubscriptionService(subs, feeds, contacts)

	err := svc.Bind(context.Background(), "feed-1", "contact-1")
	if err != nil {
		t.Fatalf("Bind returned error: %v", err)
	}

	if !subs.addCalled {
		t.Fatal("expected Add to be called")
	}
}

func TestSubscriptionServiceUnbindValidatesIDs(t *testing.T) {
	svc := NewSubscriptionService(&subsRepoStub{}, &feedLookupStub{}, &contactsRepoStub{})

	err := svc.Unbind(context.Background(), "feed-1", " ")
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

type subsRepoStub struct {
	addCalled     bool
	listByFeed    []domain.Subscription
	listByContact []domain.Subscription
}

func (s *subsRepoStub) ListByFeed(context.Context, string) ([]domain.Subscription, error) {
	return s.listByFeed, nil
}

func (s *subsRepoStub) ListByContact(context.Context, string) ([]domain.Subscription, error) {
	return s.listByContact, nil
}

func (s *subsRepoStub) Add(context.Context, string, string) error {
	s.addCalled = true
	return nil
}

func (s *subsRepoStub) Remove(context.Context, string, string) error {
	return nil
}

func (s *subsRepoStub) AddGroup(context.Context, string, string, string) error {
	return nil
}

func (s *subsRepoStub) RemoveGroupByFeed(context.Context, string, string) error {
	return nil
}

func (s *subsRepoStub) RemoveGroupByContact(context.Context, string, string) error {
	return nil
}

type feedLookupStub struct {
	feed domain.Feed
	err  error
}

func (f *feedLookupStub) Create(context.Context, domain.Feed) (domain.Feed, error) {
	return domain.Feed{}, nil
}

func (f *feedLookupStub) ListDue(context.Context, time.Time, int) ([]domain.Feed, error) {
	return nil, nil
}

func (f *feedLookupStub) MarkFetched(context.Context, string, time.Time, time.Time, *string, *string) error {
	return nil
}

func (f *feedLookupStub) MarkFetchError(context.Context, string, string) error {
	return nil
}

func (f *feedLookupStub) GetByID(context.Context, string) (domain.Feed, error) {
	if f.err != nil {
		return domain.Feed{}, f.err
	}
	return f.feed, nil
}

func (f *feedLookupStub) UpdateBatching(context.Context, string, bool, int) (domain.Feed, error) {
	return domain.Feed{}, nil
}

func (f *feedLookupStub) List(context.Context, int, int) ([]domain.Feed, int, error) {
	return nil, 0, nil
}

func (f *feedLookupStub) Delete(context.Context, string) error {
	return nil
}

type contactsRepoStub struct {
	contact domain.Contact
	err     error
}

func (c *contactsRepoStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) CreateTelegram(context.Context, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) UpdateTelegram(context.Context, string, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) CreateEmail(context.Context, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) UpdateEmail(context.Context, string, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) GetEmailConfig(context.Context, string) (domain.EmailContactConfig, error) {
	return domain.EmailContactConfig{}, nil
}

func (c *contactsRepoStub) CreateHTTP(context.Context, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) UpdateHTTP(context.Context, string, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}

func (c *contactsRepoStub) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (c *contactsRepoStub) GetByID(context.Context, string) (domain.Contact, error) {
	if c.err != nil {
		return domain.Contact{}, c.err
	}
	return c.contact, nil
}

func (c *contactsRepoStub) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return nil, 0, nil
}

func (c *contactsRepoStub) Delete(context.Context, string) error {
	return nil
}
