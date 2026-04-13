package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestContactServiceCreateTelegramValidatesInput(t *testing.T) {
	svc := NewContactService(&contactRepoStub{})

	tests := []CreateTelegramContactInput{
		{ChatID: ""},
		{ChatID: "123", Status: "weird"},
	}

	for _, in := range tests {
		_, err := svc.CreateTelegram(context.Background(), in)
		if !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for %+v, got %v", in, err)
		}
	}
}

func TestContactServiceCreateTelegramNormalizesPayload(t *testing.T) {
	repo := &contactRepoStub{createdTelegram: domain.Contact{ID: "tg-1", Type: domain.ContactTelegram}}
	svc := NewContactService(repo)

	got, err := svc.CreateTelegram(context.Background(), CreateTelegramContactInput{
		ChatID:      "12345",
		Username:    strPtr(" alice "),
		DisplayName: strPtr(" Alice "),
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("CreateTelegram returned error: %v", err)
	}

	if got.ID != "tg-1" {
		t.Fatalf("unexpected contact: %+v", got)
	}
	if repo.createTelegramChatID != "12345" || repo.createTelegramStatus != domain.ContactActive {
		t.Fatalf("unexpected repo call: %+v", repo)
	}
	if repo.createTelegramUsername == nil || *repo.createTelegramUsername != "alice" {
		t.Fatalf("username was not normalized: %+v", repo.createTelegramUsername)
	}
}

func TestContactServiceCreateHTTPValidatesInput(t *testing.T) {
	svc := NewContactService(&contactRepoStub{})

	tests := []CreateHTTPContactInput{
		{Method: "TRACE", URL: "https://example.com/hook"},
		{Method: "POST", URL: ""},
		{Method: "POST", URL: "ftp://example.com/hook"},
		{Method: "POST", URL: "://bad"},
		{Method: "POST", URL: "https://example.com/hook", Headers: map[string]string{"Bad\nHeader": "x"}},
	}

	for _, in := range tests {
		_, err := svc.CreateHTTP(context.Background(), in)
		if !errors.Is(err, ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest for %+v, got %v", in, err)
		}
	}
}

func TestContactServiceCreateHTTPNormalizesPayload(t *testing.T) {
	repo := &contactRepoStub{createdHTTP: domain.Contact{ID: "http-1", Type: domain.ContactHTTP}}
	svc := NewContactService(repo)

	got, err := svc.CreateHTTP(context.Background(), CreateHTTPContactInput{
		DisplayName:  strPtr("  Webhook A  "),
		Method:       "post",
		URL:          "https://example.com/hook",
		Headers:      map[string]string{"content-type": "application/json"},
		BodyTemplate: strPtr("{\"text\": {json_text}}"),
	})
	if err != nil {
		t.Fatalf("CreateHTTP returned error: %v", err)
	}

	if got.ID != "http-1" {
		t.Fatalf("unexpected contact: %+v", got)
	}
	if repo.createHTTP.Method != "POST" {
		t.Fatalf("expected POST, got %q", repo.createHTTP.Method)
	}
	if repo.createHTTPDisplayName == nil || *repo.createHTTPDisplayName != "Webhook A" {
		t.Fatalf("unexpected display name: %+v", repo.createHTTPDisplayName)
	}
	if repo.createHTTP.Headers["Content-Type"] != "application/json" {
		t.Fatalf("headers were not normalized: %+v", repo.createHTTP.Headers)
	}
	if repo.createHTTPStatus != domain.ContactActive {
		t.Fatalf("expected active status, got %s", repo.createHTTPStatus)
	}
}

func TestContactServiceUpdateHTTPMapsNotFound(t *testing.T) {
	repo := &contactRepoStub{updateHTTPErr: sql.ErrNoRows}
	svc := NewContactService(repo)

	_, err := svc.UpdateHTTP(context.Background(), "contact-1", UpdateHTTPContactInput{
		Status: "active",
		Method: "POST",
		URL:    "https://example.com/hook",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContactServiceDeleteMapsNotFound(t *testing.T) {
	repo := &contactRepoStub{deleteErr: sql.ErrNoRows}
	svc := NewContactService(repo)

	err := svc.Delete(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContactServiceCreateEmailNormalizesPayload(t *testing.T) {
	repo := &contactRepoStub{createdEmail: domain.Contact{ID: "email-1", Type: domain.ContactEmail}}
	svc := NewContactService(repo)

	got, err := svc.CreateEmail(context.Background(), CreateEmailContactInput{
		Email:       " Alice@Example.com ",
		DisplayName: strPtr(" Alice "),
		Format:      "html",
	})
	if err != nil {
		t.Fatalf("CreateEmail returned error: %v", err)
	}

	if got.ID != "email-1" {
		t.Fatalf("unexpected contact: %+v", got)
	}
	if repo.createEmailValue != "alice@example.com" {
		t.Fatalf("email was not normalized: %q", repo.createEmailValue)
	}
	if repo.createEmailCfg.Format != "html" {
		t.Fatalf("unexpected format: %+v", repo.createEmailCfg)
	}
}

type contactRepoStub struct {
	createdTelegram        domain.Contact
	createTelegramChatID   string
	createTelegramUsername *string
	createTelegramDisplay  *string
	createTelegramStatus   domain.ContactStatus

	createdHTTP           domain.Contact
	createHTTP            domain.HTTPContactConfig
	createHTTPDisplayName *string
	createHTTPStatus      domain.ContactStatus
	createdEmail          domain.Contact
	createEmailValue      string
	createEmailCfg        domain.EmailContactConfig
	createEmailDisplay    *string
	createEmailStatus     domain.ContactStatus

	updateHTTPErr error
	deleteErr     error
}

func (s *contactRepoStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactRepoStub) CreateTelegram(_ context.Context, chatID string, username *string, displayName *string, status domain.ContactStatus, _ *time.Time) (domain.Contact, error) {
	s.createTelegramChatID = chatID
	s.createTelegramUsername = username
	s.createTelegramDisplay = displayName
	s.createTelegramStatus = status
	return s.createdTelegram, nil
}

func (s *contactRepoStub) UpdateTelegram(context.Context, string, string, *string, *string, domain.ContactStatus, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactRepoStub) CreateEmail(_ context.Context, value string, displayName *string, status domain.ContactStatus, cfg domain.EmailContactConfig, _ *time.Time) (domain.Contact, error) {
	s.createEmailValue = value
	s.createEmailCfg = cfg
	s.createEmailDisplay = displayName
	s.createEmailStatus = status
	return s.createdEmail, nil
}

func (s *contactRepoStub) UpdateEmail(context.Context, string, string, *string, domain.ContactStatus, domain.EmailContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactRepoStub) GetEmailConfig(context.Context, string) (domain.EmailContactConfig, error) {
	return domain.EmailContactConfig{}, nil
}

func (s *contactRepoStub) CreateHTTP(_ context.Context, _ string, displayName *string, status domain.ContactStatus, cfg domain.HTTPContactConfig, _ *time.Time) (domain.Contact, error) {
	s.createHTTP = cfg
	s.createHTTPDisplayName = displayName
	s.createHTTPStatus = status
	return s.createdHTTP, nil
}

func (s *contactRepoStub) UpdateHTTP(context.Context, string, string, *string, domain.ContactStatus, domain.HTTPContactConfig, *time.Time) (domain.Contact, error) {
	return domain.Contact{}, s.updateHTTPErr
}

func (s *contactRepoStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}

func (s *contactRepoStub) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactRepoStub) GetByID(context.Context, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (s *contactRepoStub) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return nil, 0, nil
}

func (s *contactRepoStub) Delete(context.Context, string) error {
	return s.deleteErr
}

func strPtr(v string) *string { return &v }
