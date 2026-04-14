package service

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

var ErrAlreadyRegistered = errors.New("already_registered")
var (
	ErrRegistrationCodeNotFound  = errors.New("registration_code_not_found")
	ErrRegistrationCodeDisabled  = errors.New("registration_code_disabled")
	ErrRegistrationCodeExpired   = errors.New("registration_code_expired")
	ErrRegistrationCodeExhausted = errors.New("registration_code_exhausted")
)

type RegistrationService struct {
	contacts      repository.ContactsRepository
	codes         repository.RegistrationCodesRepository
	groups        repository.GroupsRepository
	subscriptions repository.SubscriptionsRepository
	txRunner      repository.Database
}

func NewRegistrationService(
	contacts repository.ContactsRepository,
	codes repository.RegistrationCodesRepository,
	groups repository.GroupsRepository,
	subscriptions repository.SubscriptionsRepository,
	txRunner ...repository.Database,
) *RegistrationService {
	var db repository.Database
	if len(txRunner) > 0 {
		db = txRunner[0]
	}
	return &RegistrationService{
		contacts:      contacts,
		codes:         codes,
		groups:        groups,
		subscriptions: subscriptions,
		txRunner:      db,
	}
}

type RegisterTelegramInput struct {
	ChatID      string
	Username    *string
	DisplayName *string
	Code        string
}

type RegisterEmailInput struct {
	Email       string
	DisplayName *string
	Code        string
	Format      string
}

type RegisterResult struct {
	Contact       domain.Contact
	AppliedCode   *domain.RegistrationCode
	AppliedGroups []domain.Group
}

func (s *RegistrationService) RegisterTelegram(ctx context.Context, in RegisterTelegramInput) (RegisterResult, error) {
	chatID := strings.TrimSpace(in.ChatID)
	if chatID == "" {
		return RegisterResult{}, ErrBadRequest
	}

	username := normalizeOptional(in.Username)
	displayName := normalizeOptional(in.DisplayName)
	now := time.Now().UTC()

	var result RegisterResult
	err := s.withRepos(ctx, func(contacts repository.ContactsRepository, codes repository.RegistrationCodesRepository, groups repository.GroupsRepository, subscriptions repository.SubscriptionsRepository) error {
		contact, err := s.createContactIfNeeded(
			ctx,
			contacts,
			domain.ContactTelegram,
			chatID,
			func() (domain.Contact, error) {
				return contacts.CreateTelegram(ctx, chatID, username, displayName, domain.ContactActive, &now)
			},
		)
		if err != nil {
			return err
		}

		result, err = s.applyRegistrationCode(ctx, codes, groups, subscriptions, contact, in.Code, now)
		return err
	})
	return result, err
}

func (s *RegistrationService) RegisterEmail(ctx context.Context, in RegisterEmailInput) (RegisterResult, error) {
	emailValue := strings.ToLower(strings.TrimSpace(in.Email))
	if emailValue == "" {
		return RegisterResult{}, ErrBadRequest
	}

	parsed, err := mail.ParseAddress(emailValue)
	if err != nil || !strings.EqualFold(parsed.Address, emailValue) {
		return RegisterResult{}, ErrBadRequest
	}

	displayName := normalizeOptional(in.DisplayName)
	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format == "" {
		format = "plain"
	}
	now := time.Now().UTC()

	var result RegisterResult
	err = s.withRepos(ctx, func(contacts repository.ContactsRepository, codes repository.RegistrationCodesRepository, groups repository.GroupsRepository, subscriptions repository.SubscriptionsRepository) error {
		contact, err := s.createContactIfNeeded(
			ctx,
			contacts,
			domain.ContactEmail,
			emailValue,
			func() (domain.Contact, error) {
				return contacts.CreateEmail(ctx, emailValue, displayName, domain.ContactActive, domain.EmailContactConfig{Format: format}, &now)
			},
		)
		if err != nil {
			return err
		}

		result, err = s.applyRegistrationCode(ctx, codes, groups, subscriptions, contact, in.Code, now)
		return err
	})
	return result, err
}

func (s *RegistrationService) createContactIfNeeded(
	ctx context.Context,
	contacts repository.ContactsRepository,
	contactType domain.ContactType,
	value string,
	create func() (domain.Contact, error),
) (domain.Contact, error) {
	if existing, err := contacts.GetByTypeValue(ctx, contactType, value); err == nil {
		if existing.Status == domain.ContactActive {
			return domain.Contact{}, ErrAlreadyRegistered
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, err
	}

	return create()
}

func (s *RegistrationService) applyRegistrationCode(ctx context.Context, codes repository.RegistrationCodesRepository, groups repository.GroupsRepository, subscriptions repository.SubscriptionsRepository, contact domain.Contact, codeInput string, now time.Time) (RegisterResult, error) {
	result := RegisterResult{Contact: contact}

	codeValue := strings.ToUpper(strings.TrimSpace(codeInput))
	if codeValue == "" {
		return result, nil
	}

	code, err := codes.GetByCode(ctx, codeValue)
	if errors.Is(err, sql.ErrNoRows) {
		return RegisterResult{}, ErrRegistrationCodeNotFound
	}
	if err != nil {
		return RegisterResult{}, err
	}
	if !code.Enabled {
		return RegisterResult{}, ErrRegistrationCodeDisabled
	}
	if code.ExpiresAt != nil && code.ExpiresAt.Before(now) {
		return RegisterResult{}, ErrRegistrationCodeExpired
	}
	if code.MaxUses != nil && code.UseCount >= *code.MaxUses {
		return RegisterResult{}, ErrRegistrationCodeExhausted
	}

	appliedGroups, err := codes.ListGroups(ctx, code.ID)
	if err != nil {
		return RegisterResult{}, err
	}
	for _, group := range appliedGroups {
		if err := groups.AddContact(ctx, group.ID, contact.ID); err != nil {
			return RegisterResult{}, err
		}
		feedIDs, err := groups.ListFeedIDs(ctx, group.ID)
		if err != nil {
			return RegisterResult{}, err
		}
		for _, feedID := range feedIDs {
			if err := subscriptions.AddGroup(ctx, feedID, contact.ID, group.ID); err != nil {
				return RegisterResult{}, err
			}
		}
	}
	if err := codes.IncrementUse(ctx, code.ID); err != nil {
		return RegisterResult{}, err
	}

	result.AppliedCode = &code
	result.AppliedGroups = appliedGroups
	return result, nil
}

func (s *RegistrationService) withRepos(ctx context.Context, fn func(repository.ContactsRepository, repository.RegistrationCodesRepository, repository.GroupsRepository, repository.SubscriptionsRepository) error) error {
	if s.txRunner == nil {
		return fn(s.contacts, s.codes, s.groups, s.subscriptions)
	}
	return s.txRunner.WithinTx(ctx, func(store repository.Store) error {
		return fn(store.Contacts(), store.RegistrationCodes(), store.Groups(), store.Subscriptions())
	})
}
