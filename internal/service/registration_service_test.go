package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestRegistrationServiceRegisterTelegramWithCode(t *testing.T) {
	contacts := &registrationContactsRepoStub{
		getByTypeErr: sql.ErrNoRows,
		created:      domain.Contact{ID: "contact-1", Type: domain.ContactTelegram, Status: domain.ContactActive},
	}
	codes := &registrationCodesRepoForRegistrationStub{
		code: domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true},
		groups: []domain.Group{
			{ID: "group-1", Name: "vip"},
		},
	}
	groups := &registrationGroupsRepoStub{feedIDs: map[string][]string{"group-1": {"feed-1"}}}
	subs := &subscriptionsRepoStubForGroups{}
	svc := NewRegistrationService(contacts, codes, groups, subs)

	got, err := svc.RegisterTelegram(context.Background(), RegisterTelegramInput{
		ChatID: "12345",
		Code:   "abc123",
	})
	if err != nil {
		t.Fatalf("RegisterTelegram returned error: %v", err)
	}
	if got.AppliedCode == nil || got.AppliedCode.Code != "ABC123" {
		t.Fatalf("expected applied code, got %+v", got)
	}
	if len(got.AppliedGroups) != 1 || got.AppliedGroups[0].Name != "vip" {
		t.Fatalf("unexpected groups: %+v", got.AppliedGroups)
	}
	if !codes.incremented {
		t.Fatal("expected code usage to be incremented")
	}
	if len(subs.added) != 1 || subs.added[0] != [3]string{"feed-1", "contact-1", "group-1"} {
		t.Fatalf("unexpected materialized subscriptions: %+v", subs.added)
	}
}

func TestRegistrationServiceRejectsAlreadyRegisteredTelegramContact(t *testing.T) {
	svc := NewRegistrationService(
		&registrationContactsRepoStub{
			existing: domain.Contact{ID: "contact-1", Type: domain.ContactTelegram, Status: domain.ContactActive},
		},
		&registrationCodesRepoForRegistrationStub{},
		&registrationGroupsRepoStub{},
		&subscriptionsRepoStubForGroups{},
	)

	_, err := svc.RegisterTelegram(context.Background(), RegisterTelegramInput{ChatID: "12345"})
	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestRegistrationServiceRejectsExpiredRegistrationCode(t *testing.T) {
	expired := time.Now().UTC().Add(-time.Hour)
	svc := NewRegistrationService(
		&registrationContactsRepoStub{
			getByTypeErr: sql.ErrNoRows,
			created:      domain.Contact{ID: "contact-1", Type: domain.ContactTelegram, Status: domain.ContactActive},
		},
		&registrationCodesRepoForRegistrationStub{
			code: domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true, ExpiresAt: &expired},
		},
		&registrationGroupsRepoStub{},
		&subscriptionsRepoStubForGroups{},
	)

	_, err := svc.RegisterTelegram(context.Background(), RegisterTelegramInput{ChatID: "12345", Code: "ABC123"})
	if !errors.Is(err, ErrRegistrationCodeExpired) {
		t.Fatalf("expected ErrRegistrationCodeExpired, got %v", err)
	}
}

func TestRegistrationServiceRejectsExhaustedRegistrationCode(t *testing.T) {
	maxUses := 1
	svc := NewRegistrationService(
		&registrationContactsRepoStub{
			getByTypeErr: sql.ErrNoRows,
			created:      domain.Contact{ID: "contact-1", Type: domain.ContactTelegram, Status: domain.ContactActive},
		},
		&registrationCodesRepoForRegistrationStub{
			code: domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true, MaxUses: &maxUses, UseCount: 1},
		},
		&registrationGroupsRepoStub{},
		&subscriptionsRepoStubForGroups{},
	)

	_, err := svc.RegisterTelegram(context.Background(), RegisterTelegramInput{ChatID: "12345", Code: "ABC123"})
	if !errors.Is(err, ErrRegistrationCodeExhausted) {
		t.Fatalf("expected ErrRegistrationCodeExhausted, got %v", err)
	}
}

func TestRegistrationServiceRegisterEmailWithCode(t *testing.T) {
	contacts := &registrationContactsRepoStub{
		getByTypeErr: sql.ErrNoRows,
		createdEmail: domain.Contact{ID: "contact-email-1", Type: domain.ContactEmail, Status: domain.ContactActive, Value: "alice@example.com"},
	}
	codes := &registrationCodesRepoForRegistrationStub{
		code: domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true},
	}
	svc := NewRegistrationService(contacts, codes, &registrationGroupsRepoStub{}, &subscriptionsRepoStubForGroups{})

	got, err := svc.RegisterEmail(context.Background(), RegisterEmailInput{
		Email: "Alice@Example.com",
		Code:  "abc123",
	})
	if err != nil {
		t.Fatalf("RegisterEmail returned error: %v", err)
	}
	if got.Contact.Value != "alice@example.com" {
		t.Fatalf("expected normalized email contact, got %+v", got.Contact)
	}
	if contacts.createEmailValue != "alice@example.com" {
		t.Fatalf("expected normalized email to be persisted, got %q", contacts.createEmailValue)
	}
}

type registrationContactsRepoStub struct {
	existing         domain.Contact
	getByTypeErr     error
	created          domain.Contact
	createdEmail     domain.Contact
	createEmailValue string
}

func (s *registrationContactsRepoStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) CreateTelegram(context.Context, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return s.created, nil
}
func (s *registrationContactsRepoStub) UpdateTelegram(context.Context, string, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) CreateEmail(_ context.Context, value string, _ *string, _ domain.ContactStatus, _ domain.EmailContactConfig, _ *time.Time) (domain.Contact, error) {
	s.createEmailValue = value
	return s.createdEmail, nil
}
func (s *registrationContactsRepoStub) UpdateEmail(context.Context, string, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) GetEmailConfig(context.Context, string) (domain.EmailContactConfig, error) {
	return domain.EmailContactConfig{}, nil
}
func (s *registrationContactsRepoStub) CreateHTTP(context.Context, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) UpdateHTTP(context.Context, string, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}
func (s *registrationContactsRepoStub) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	if s.getByTypeErr != nil {
		return domain.Contact{}, s.getByTypeErr
	}
	return s.existing, nil
}
func (s *registrationContactsRepoStub) GetByID(context.Context, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (s *registrationContactsRepoStub) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return nil, 0, nil
}
func (s *registrationContactsRepoStub) Delete(context.Context, string) error { return nil }

type registrationCodesRepoForRegistrationStub struct {
	code        domain.RegistrationCode
	groups      []domain.Group
	incremented bool
}

func (s *registrationCodesRepoForRegistrationStub) Create(context.Context, domain.RegistrationCode) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (s *registrationCodesRepoForRegistrationStub) Update(context.Context, string, string, string, *string, bool, *int, *time.Time) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (s *registrationCodesRepoForRegistrationStub) GetByID(context.Context, string) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (s *registrationCodesRepoForRegistrationStub) GetByCode(context.Context, string) (domain.RegistrationCode, error) {
	return s.code, nil
}
func (s *registrationCodesRepoForRegistrationStub) List(context.Context, int, int) ([]domain.RegistrationCode, int, error) {
	return nil, 0, nil
}
func (s *registrationCodesRepoForRegistrationStub) Delete(context.Context, string) error { return nil }
func (s *registrationCodesRepoForRegistrationStub) ListGroups(context.Context, string) ([]domain.Group, error) {
	return s.groups, nil
}
func (s *registrationCodesRepoForRegistrationStub) AddGroup(context.Context, string, string) error {
	return nil
}
func (s *registrationCodesRepoForRegistrationStub) RemoveGroup(context.Context, string, string) error {
	return nil
}
func (s *registrationCodesRepoForRegistrationStub) IncrementUse(context.Context, string) error {
	s.incremented = true
	return nil
}

type registrationGroupsRepoStub struct {
	feedIDs map[string][]string
}

func (s *registrationGroupsRepoStub) Create(context.Context, domain.Group) (domain.Group, error) {
	return domain.Group{}, nil
}
func (s *registrationGroupsRepoStub) Update(context.Context, string, string, *string) (domain.Group, error) {
	return domain.Group{}, nil
}
func (s *registrationGroupsRepoStub) GetByID(context.Context, string) (domain.Group, error) {
	return domain.Group{}, nil
}
func (s *registrationGroupsRepoStub) List(context.Context, int, int) ([]domain.Group, int, error) {
	return nil, 0, nil
}
func (s *registrationGroupsRepoStub) Delete(context.Context, string) error { return nil }
func (s *registrationGroupsRepoStub) ListContacts(context.Context, string) ([]domain.Contact, error) {
	return nil, nil
}
func (s *registrationGroupsRepoStub) AddContact(context.Context, string, string) error { return nil }
func (s *registrationGroupsRepoStub) RemoveContact(context.Context, string, string) error {
	return nil
}
func (s *registrationGroupsRepoStub) ListFeeds(context.Context, string) ([]domain.Feed, error) {
	return nil, nil
}
func (s *registrationGroupsRepoStub) AddFeed(context.Context, string, string) error    { return nil }
func (s *registrationGroupsRepoStub) RemoveFeed(context.Context, string, string) error { return nil }
func (s *registrationGroupsRepoStub) ListFeedIDs(_ context.Context, groupID string) ([]string, error) {
	return s.feedIDs[groupID], nil
}
func (s *registrationGroupsRepoStub) ListContactIDs(context.Context, string) ([]string, error) {
	return nil, nil
}
