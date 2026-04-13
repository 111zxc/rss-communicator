package service

import (
	"context"
	"database/sql"
	"errors"
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
}

func NewRegistrationService(
	contacts repository.ContactsRepository,
	codes repository.RegistrationCodesRepository,
	groups repository.GroupsRepository,
	subscriptions repository.SubscriptionsRepository,
) *RegistrationService {
	return &RegistrationService{
		contacts:      contacts,
		codes:         codes,
		groups:        groups,
		subscriptions: subscriptions,
	}
}

type RegisterTelegramInput struct {
	ChatID      string
	Username    *string
	DisplayName *string
	Code        string
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

	if existing, err := s.contacts.GetByTypeValue(ctx, domain.ContactTelegram, chatID); err == nil {
		if existing.Status == domain.ContactActive {
			return RegisterResult{}, ErrAlreadyRegistered
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return RegisterResult{}, err
	}

	username := normalizeOptional(in.Username)
	displayName := normalizeOptional(in.DisplayName)
	now := time.Now().UTC()

	contact, err := s.contacts.CreateTelegram(ctx, chatID, username, displayName, domain.ContactActive, &now)
	if err != nil {
		return RegisterResult{}, err
	}

	result := RegisterResult{Contact: contact}

	codeValue := strings.ToUpper(strings.TrimSpace(in.Code))
	if codeValue == "" {
		return result, nil
	}

	code, err := s.codes.GetByCode(ctx, codeValue)
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

	groups, err := s.codes.ListGroups(ctx, code.ID)
	if err != nil {
		return RegisterResult{}, err
	}
	for _, group := range groups {
		if err := s.groups.AddContact(ctx, group.ID, contact.ID); err != nil {
			return RegisterResult{}, err
		}
		feedIDs, err := s.groups.ListFeedIDs(ctx, group.ID)
		if err != nil {
			return RegisterResult{}, err
		}
		for _, feedID := range feedIDs {
			if err := s.subscriptions.AddGroup(ctx, feedID, contact.ID, group.ID); err != nil {
				return RegisterResult{}, err
			}
		}
	}
	if err := s.codes.IncrementUse(ctx, code.ID); err != nil {
		return RegisterResult{}, err
	}

	result.AppliedCode = &code
	result.AppliedGroups = groups
	return result, nil
}
