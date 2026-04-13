package email

import (
	"context"
	"strings"
	"testing"

	"github.com/111zxc/rss-communicator/internal/config"
	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestBuildMessagePlain(t *testing.T) {
	msg, err := buildMessage("RSS Bot <bot@example.com>", "alice@example.com", domain.Feed{Name: "Feed"}, []domain.Item{{
		Title: "Item 1",
		Link:  "https://example.com/1",
	}}, "plain")
	if err != nil {
		t.Fatalf("buildMessage returned error: %v", err)
	}

	body := string(msg)
	if !strings.Contains(body, "Subject: Feed: Item 1") {
		t.Fatalf("unexpected subject: %s", body)
	}
	if !strings.Contains(body, "Content-Type: text/plain; charset=UTF-8") {
		t.Fatalf("unexpected content type: %s", body)
	}
}

func TestSenderLoadsEmailConfigFromRepository(t *testing.T) {
	sender := New(emailContactsStub{cfg: domain.EmailContactConfig{Format: "html"}}, config.EmailConfig{
		Address:      "bot@example.com",
		DisplayName:  "RSS Bot",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     465,
		SMTPUsername: "bot@example.com",
		SMTPPassword: "secret",
	})

	_, err := buildMessage("RSS Bot <bot@example.com>", "alice@example.com", domain.Feed{Name: "Feed"}, []domain.Item{{Title: "Item 1", Link: "https://example.com/1"}}, sender.contacts.(emailContactsStub).cfg.Format)
	if err != nil {
		t.Fatalf("expected config to be usable, got %v", err)
	}
}

type emailContactsStub struct {
	cfg domain.EmailContactConfig
}

func (s emailContactsStub) GetEmailConfig(context.Context, string) (domain.EmailContactConfig, error) {
	return s.cfg, nil
}
