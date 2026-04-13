package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/repository"
)

type ContactService struct {
	contacts repository.ContactsRepository
}

func NewContactService(contacts repository.ContactsRepository) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) List(ctx context.Context, limit, offset int) ([]domain.Contact, int, error) {
	return s.contacts.List(ctx, limit, offset)
}

func (s *ContactService) GetByID(ctx context.Context, contactID string) (domain.Contact, error) {
	if strings.TrimSpace(contactID) == "" {
		return domain.Contact{}, ErrBadRequest
	}
	contact, err := s.contacts.GetByID(ctx, contactID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

type CreateTelegramContactInput struct {
	ChatID      string
	Username    *string
	DisplayName *string
	Status      string
}

type UpdateTelegramContactInput struct {
	ChatID      string
	Username    *string
	DisplayName *string
	Status      string
}

type CreateEmailContactInput struct {
	Email       string
	DisplayName *string
	Status      string
	Format      string
}

type UpdateEmailContactInput struct {
	Email       string
	DisplayName *string
	Status      string
	Format      string
}

func (s *ContactService) CreateTelegram(ctx context.Context, in CreateTelegramContactInput) (domain.Contact, error) {
	chatID := strings.TrimSpace(in.ChatID)
	if chatID == "" {
		return domain.Contact{}, ErrBadRequest
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	username := normalizeOptional(in.Username)
	displayName := normalizeOptional(in.DisplayName)
	contact, err := s.contacts.CreateTelegram(ctx, chatID, username, displayName, status, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

func (s *ContactService) CreateEmail(ctx context.Context, in CreateEmailContactInput) (domain.Contact, error) {
	emailValue, cfg, displayName, err := normalizeEmailContactInput(in.Email, in.DisplayName, in.Format)
	if err != nil {
		return domain.Contact{}, err
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	contact, err := s.contacts.CreateEmail(ctx, emailValue, displayName, status, cfg, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

func (s *ContactService) UpdateTelegram(ctx context.Context, contactID string, in UpdateTelegramContactInput) (domain.Contact, error) {
	if strings.TrimSpace(contactID) == "" {
		return domain.Contact{}, ErrBadRequest
	}
	chatID := strings.TrimSpace(in.ChatID)
	if chatID == "" {
		return domain.Contact{}, ErrBadRequest
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	username := normalizeOptional(in.Username)
	displayName := normalizeOptional(in.DisplayName)
	contact, err := s.contacts.UpdateTelegram(ctx, contactID, chatID, username, displayName, status, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

func (s *ContactService) UpdateEmail(ctx context.Context, contactID string, in UpdateEmailContactInput) (domain.Contact, error) {
	if strings.TrimSpace(contactID) == "" {
		return domain.Contact{}, ErrBadRequest
	}

	emailValue, cfg, displayName, err := normalizeEmailContactInput(in.Email, in.DisplayName, in.Format)
	if err != nil {
		return domain.Contact{}, err
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	contact, err := s.contacts.UpdateEmail(ctx, contactID, emailValue, displayName, status, cfg, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

type CreateHTTPContactInput struct {
	DisplayName  *string
	Status       string
	Method       string
	URL          string
	Headers      map[string]string
	BodyTemplate *string
}

func (s *ContactService) CreateHTTP(ctx context.Context, in CreateHTTPContactInput) (domain.Contact, error) {
	cfg, displayName, err := normalizeHTTPContactInput(in)
	if err != nil {
		return domain.Contact{}, err
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	contact, err := s.contacts.CreateHTTP(ctx, cfg.URL, displayName, status, cfg, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

type UpdateHTTPContactInput struct {
	DisplayName  *string
	Status       string
	Method       string
	URL          string
	Headers      map[string]string
	BodyTemplate *string
}

func (s *ContactService) UpdateHTTP(ctx context.Context, contactID string, in UpdateHTTPContactInput) (domain.Contact, error) {
	if strings.TrimSpace(contactID) == "" {
		return domain.Contact{}, ErrBadRequest
	}

	cfg, displayName, err := normalizeHTTPContactInput(CreateHTTPContactInput{
		DisplayName:  in.DisplayName,
		Status:       in.Status,
		Method:       in.Method,
		URL:          in.URL,
		Headers:      in.Headers,
		BodyTemplate: in.BodyTemplate,
	})
	if err != nil {
		return domain.Contact{}, err
	}

	status, verifiedAt, err := normalizeContactStatus(in.Status)
	if err != nil {
		return domain.Contact{}, err
	}

	contact, err := s.contacts.UpdateHTTP(ctx, contactID, cfg.URL, displayName, status, cfg, verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, ErrNotFound
	}
	return contact, err
}

func (s *ContactService) Delete(ctx context.Context, contactID string) error {
	if strings.TrimSpace(contactID) == "" {
		return ErrBadRequest
	}
	err := s.contacts.Delete(ctx, contactID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeEmailContactInput(emailValue string, displayName *string, format string) (string, domain.EmailContactConfig, *string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(emailValue))
	if normalizedEmail == "" {
		return "", domain.EmailContactConfig{}, nil, ErrBadRequest
	}

	parsed, err := mail.ParseAddress(normalizedEmail)
	if err != nil || !strings.EqualFold(parsed.Address, normalizedEmail) {
		return "", domain.EmailContactConfig{}, nil, ErrBadRequest
	}

	normalizedFormat := strings.ToLower(strings.TrimSpace(format))
	if normalizedFormat == "" {
		normalizedFormat = "plain"
	}
	switch normalizedFormat {
	case "plain", "html":
	default:
		return "", domain.EmailContactConfig{}, nil, ErrBadRequest
	}

	return normalizedEmail, domain.EmailContactConfig{Format: normalizedFormat}, normalizeOptional(displayName), nil
}

func normalizeHTTPContactInput(in CreateHTTPContactInput) (domain.HTTPContactConfig, *string, error) {
	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if method == "" {
		method = http.MethodPost
	}

	switch method {
	case http.MethodDelete, http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodPut:
	default:
		return domain.HTTPContactConfig{}, nil, ErrBadRequest
	}

	rawURL := strings.TrimSpace(in.URL)
	if rawURL == "" {
		return domain.HTTPContactConfig{}, nil, ErrBadRequest
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return domain.HTTPContactConfig{}, nil, ErrBadRequest
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return domain.HTTPContactConfig{}, nil, ErrBadRequest
	}

	displayName := normalizeOptional(in.DisplayName)

	headers := make(map[string]string, len(in.Headers))
	for k, v := range in.Headers {
		key := strings.TrimSpace(k)
		if key == "" || strings.ContainsAny(key, "\r\n:") {
			return domain.HTTPContactConfig{}, nil, ErrBadRequest
		}
		if strings.ContainsAny(v, "\r\n") {
			return domain.HTTPContactConfig{}, nil, ErrBadRequest
		}
		headers[http.CanonicalHeaderKey(key)] = v
	}

	var bodyTemplate *string
	if in.BodyTemplate != nil {
		body := *in.BodyTemplate
		bodyTemplate = &body
	}

	return domain.HTTPContactConfig{
		Method:       method,
		URL:          rawURL,
		Headers:      headers,
		BodyTemplate: bodyTemplate,
	}, displayName, nil
}

func normalizeOptional(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeContactStatus(raw string) (domain.ContactStatus, *time.Time, error) {
	status := domain.ContactStatus(strings.TrimSpace(strings.ToLower(raw)))
	if status == "" {
		status = domain.ContactActive
	}
	switch status {
	case domain.ContactActive:
		now := time.Now().UTC()
		return status, &now, nil
	case domain.ContactPending, domain.ContactDisabled:
		return status, nil, nil
	default:
		return "", nil, ErrBadRequest
	}
}
