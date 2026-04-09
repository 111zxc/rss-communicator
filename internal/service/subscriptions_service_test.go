package service

import (
	"context"
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
	addCalled bool
}

func (s *subsRepoStub) ListByFeed(context.Context, string) ([]domain.Subscription, error) {
	return nil, nil
}

func (s *subsRepoStub) Add(context.Context, string, string) error {
	s.addCalled = true
	return nil
}

func (s *subsRepoStub) Remove(context.Context, string, string) error {
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
