package service

import (
	"context"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestGroupServiceAddContactMaterializesGroupSubscriptions(t *testing.T) {
	groups := &groupRepoStub{
		group:   domain.Group{ID: "group-1", Name: "g"},
		feedIDs: []string{"feed-1", "feed-2"},
	}
	contacts := &contactsRepoStubForGroups{contact: domain.Contact{ID: "contact-1"}}
	subs := &subscriptionsRepoStubForGroups{}
	svc := NewGroupService(groups, &feedsRepoStubForGroups{}, contacts, subs)

	err := svc.AddContact(context.Background(), "group-1", "contact-1")
	if err != nil {
		t.Fatalf("AddContact returned error: %v", err)
	}
	if len(subs.added) != 2 {
		t.Fatalf("expected 2 materialized subscriptions, got %+v", subs.added)
	}
}

func TestGroupServiceAddFeedMaterializesGroupSubscriptions(t *testing.T) {
	groups := &groupRepoStub{
		group:      domain.Group{ID: "group-1", Name: "g"},
		contactIDs: []string{"contact-1", "contact-2"},
	}
	subs := &subscriptionsRepoStubForGroups{}
	svc := NewGroupService(groups, &feedsRepoStubForGroups{feed: domain.Feed{ID: "feed-1"}}, &contactsRepoStubForGroups{}, subs)

	err := svc.AddFeed(context.Background(), "group-1", "feed-1")
	if err != nil {
		t.Fatalf("AddFeed returned error: %v", err)
	}
	if len(subs.added) != 2 {
		t.Fatalf("expected 2 materialized subscriptions, got %+v", subs.added)
	}
}

type groupRepoStub struct {
	group      domain.Group
	feedIDs    []string
	contactIDs []string
}

func (s *groupRepoStub) Create(context.Context, domain.Group) (domain.Group, error) {
	return s.group, nil
}
func (s *groupRepoStub) Update(context.Context, string, string, *string) (domain.Group, error) {
	return s.group, nil
}
func (s *groupRepoStub) GetByID(context.Context, string) (domain.Group, error) { return s.group, nil }
func (s *groupRepoStub) List(context.Context, int, int) ([]domain.Group, int, error) {
	return []domain.Group{s.group}, 1, nil
}
func (s *groupRepoStub) Delete(context.Context, string) error { return nil }
func (s *groupRepoStub) ListContacts(context.Context, string) ([]domain.Contact, error) {
	return nil, nil
}
func (s *groupRepoStub) AddContact(context.Context, string, string) error         { return nil }
func (s *groupRepoStub) RemoveContact(context.Context, string, string) error      { return nil }
func (s *groupRepoStub) ListFeeds(context.Context, string) ([]domain.Feed, error) { return nil, nil }
func (s *groupRepoStub) AddFeed(context.Context, string, string) error            { return nil }
func (s *groupRepoStub) RemoveFeed(context.Context, string, string) error         { return nil }
func (s *groupRepoStub) ListFeedIDs(context.Context, string) ([]string, error)    { return s.feedIDs, nil }
func (s *groupRepoStub) ListContactIDs(context.Context, string) ([]string, error) {
	return s.contactIDs, nil
}

type feedsRepoStubForGroups struct{ feed domain.Feed }

func (s *feedsRepoStubForGroups) Create(context.Context, domain.Feed) (domain.Feed, error) {
	return domain.Feed{}, nil
}
func (s *feedsRepoStubForGroups) ListDue(context.Context, time.Time, int) ([]domain.Feed, error) {
	return nil, nil
}
func (s *feedsRepoStubForGroups) MarkFetched(context.Context, string, time.Time, time.Time, *string, *string) error {
	return nil
}
func (s *feedsRepoStubForGroups) MarkFetchError(context.Context, string, string) error { return nil }
func (s *feedsRepoStubForGroups) GetByID(context.Context, string) (domain.Feed, error) {
	return s.feed, nil
}
func (s *feedsRepoStubForGroups) UpdateBatching(context.Context, string, bool, int) (domain.Feed, error) {
	return domain.Feed{}, nil
}
func (s *feedsRepoStubForGroups) List(context.Context, int, int) ([]domain.Feed, int, error) {
	return nil, 0, nil
}
func (s *feedsRepoStubForGroups) Delete(context.Context, string) error { return nil }

type contactsRepoStubForGroups struct{ contact domain.Contact }

func (s *contactsRepoStubForGroups) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) CreateTelegram(context.Context, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) UpdateTelegram(context.Context, string, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) CreateEmail(context.Context, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) UpdateEmail(context.Context, string, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) GetEmailConfig(context.Context, string) (domain.EmailContactConfig, error) {
	return domain.EmailContactConfig{}, nil
}
func (s *contactsRepoStubForGroups) CreateHTTP(context.Context, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) UpdateHTTP(context.Context, string, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}
func (s *contactsRepoStubForGroups) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *contactsRepoStubForGroups) GetByID(context.Context, string) (domain.Contact, error) {
	return s.contact, nil
}
func (s *contactsRepoStubForGroups) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return nil, 0, nil
}
func (s *contactsRepoStubForGroups) Delete(context.Context, string) error { return nil }

type subscriptionsRepoStubForGroups struct {
	added [][3]string
}

func (s *subscriptionsRepoStubForGroups) ListByFeed(context.Context, string) ([]domain.Subscription, error) {
	return nil, nil
}
func (s *subscriptionsRepoStubForGroups) ListByContact(context.Context, string) ([]domain.Subscription, error) {
	return nil, nil
}
func (s *subscriptionsRepoStubForGroups) Add(context.Context, string, string) error { return nil }
func (s *subscriptionsRepoStubForGroups) Remove(context.Context, string, string) error {
	return nil
}
func (s *subscriptionsRepoStubForGroups) AddGroup(_ context.Context, feedID, contactID, groupID string) error {
	s.added = append(s.added, [3]string{feedID, contactID, groupID})
	return nil
}
func (s *subscriptionsRepoStubForGroups) RemoveGroupByFeed(context.Context, string, string) error {
	return nil
}
func (s *subscriptionsRepoStubForGroups) RemoveGroupByContact(context.Context, string, string) error {
	return nil
}
