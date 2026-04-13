package email

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/service"
)

type IncomingMessage struct {
	UID     uint32
	From    string
	Subject string
}

type Mailbox interface {
	FetchUnread(ctx context.Context) ([]IncomingMessage, error)
	MarkSeen(ctx context.Context, uids []uint32) error
}

type RegistrationService interface {
	RegisterEmail(ctx context.Context, in service.RegisterEmailInput) (service.RegisterResult, error)
}

type Poller struct {
	mailbox  Mailbox
	registry RegistrationService
	log      *slog.Logger
	interval time.Duration
}

func NewPoller(mailbox Mailbox, registry RegistrationService, log *slog.Logger, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Poller{mailbox: mailbox, registry: registry, log: log, interval: interval}
}

func (p *Poller) Run(ctx context.Context) error {
	t := time.NewTicker(p.interval)
	defer t.Stop()

	for {
		if err := p.pollOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			p.log.Error("email poll failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) error {
	msgs, err := p.mailbox.FetchUnread(ctx)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	seen := make([]uint32, 0, len(msgs))
	for _, msg := range msgs {
		if err := p.handleMessage(ctx, msg); err != nil {
			p.log.Warn("email registration message ignored", "uid", msg.UID, "err", err)
		}
		seen = append(seen, msg.UID)
	}

	return p.mailbox.MarkSeen(ctx, seen)
}

func (p *Poller) handleMessage(ctx context.Context, msg IncomingMessage) error {
	fromAddr, err := parseFromAddress(msg.From)
	if err != nil {
		return err
	}

	cmd, code, ok := ParseSubject(msg.Subject)
	if !ok || cmd != "CONFIRM" {
		return errors.New("unsupported subject")
	}

	_, err = p.registry.RegisterEmail(ctx, service.RegisterEmailInput{
		Email: fromAddr,
		Code:  code,
	})
	if errors.Is(err, service.ErrAlreadyRegistered) {
		return nil
	}
	return err
}

func ParseSubject(subject string) (command string, code string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(subject))
	if len(fields) == 0 {
		return "", "", false
	}
	if !strings.EqualFold(fields[0], "CONFIRM") {
		return "", "", false
	}
	if len(fields) > 2 {
		return "", "", false
	}
	if len(fields) == 2 {
		return "CONFIRM", fields[1], true
	}
	return "CONFIRM", "", true
}

func parseFromAddress(raw string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	return strings.ToLower(addr.Address), nil
}
