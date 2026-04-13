package email

import (
	"context"
	"log/slog"
	"testing"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/service"
)

func TestParseSubject(t *testing.T) {
	tests := []struct {
		subject string
		code    string
		ok      bool
	}{
		{subject: "CONFIRM", ok: true},
		{subject: " confirm ABC123 ", code: "ABC123", ok: true},
		{subject: "HELLO", ok: false},
		{subject: "CONFIRM A B", ok: false},
	}

	for _, tt := range tests {
		cmd, code, ok := ParseSubject(tt.subject)
		if ok != tt.ok || (ok && (cmd != "CONFIRM" || code != tt.code)) {
			t.Fatalf("unexpected parse result for %q: cmd=%q code=%q ok=%v", tt.subject, cmd, code, ok)
		}
	}
}

func TestPollerProcessesUnreadMessages(t *testing.T) {
	mailbox := &mailboxStub{
		msgs: []IncomingMessage{{UID: 1, From: "Alice <alice@example.com>", Subject: "CONFIRM CODE1"}},
	}
	registry := &registrationStub{}
	poller := NewPoller(mailbox, registry, slog.Default(), 0)

	if err := poller.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}
	if registry.input.Email != "alice@example.com" || registry.input.Code != "CODE1" {
		t.Fatalf("unexpected registration input: %+v", registry.input)
	}
	if len(mailbox.seen) != 1 || mailbox.seen[0] != 1 {
		t.Fatalf("expected uid 1 to be marked seen, got %+v", mailbox.seen)
	}
}

func TestPollerIgnoresAlreadyRegistered(t *testing.T) {
	mailbox := &mailboxStub{
		msgs: []IncomingMessage{{UID: 7, From: "bob@example.com", Subject: "CONFIRM"}},
	}
	registry := &registrationStub{err: service.ErrAlreadyRegistered}
	poller := NewPoller(mailbox, registry, slog.Default(), 0)

	if err := poller.pollOnce(context.Background()); err != nil {
		t.Fatalf("pollOnce returned error: %v", err)
	}
}

func TestParseFromAddressRejectsInvalidValue(t *testing.T) {
	if _, err := parseFromAddress("not-an-email"); err == nil {
		t.Fatal("expected invalid sender to fail")
	}
}

type mailboxStub struct {
	msgs []IncomingMessage
	seen []uint32
	err  error
}

func (m *mailboxStub) FetchUnread(context.Context) ([]IncomingMessage, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.msgs, nil
}

func (m *mailboxStub) MarkSeen(_ context.Context, uids []uint32) error {
	m.seen = append(m.seen, uids...)
	return nil
}

type registrationStub struct {
	input service.RegisterEmailInput
	err   error
}

func (r *registrationStub) RegisterEmail(_ context.Context, in service.RegisterEmailInput) (service.RegisterResult, error) {
	r.input = in
	if r.err != nil {
		return service.RegisterResult{}, r.err
	}
	return service.RegisterResult{Contact: domain.Contact{ID: "contact-1"}}, nil
}
