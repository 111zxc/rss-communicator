package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/111zxc/rss-communicator/internal/config"
	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
	"github.com/111zxc/rss-communicator/internal/senders/render"
)

type ContactsRepository interface {
	GetEmailConfig(ctx context.Context, contactID string) (domain.EmailContactConfig, error)
}

type Sender struct {
	contacts ContactsRepository
	cfg      config.EmailConfig
}

func New(contacts ContactsRepository, cfg config.EmailConfig) *Sender {
	return &Sender{contacts: contacts, cfg: cfg}
}

func (s *Sender) Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	if len(items) == 0 {
		return nil
	}

	cfg := c.Email
	if cfg == nil {
		loaded, err := s.contacts.GetEmailConfig(ctx, c.ID)
		if err != nil {
			return &worker.PermanentError{Msg: fmt.Sprintf("email config not found: %v", err)}
		}
		cfg = &loaded
	}

	fromHeader := s.cfg.Address
	if displayName := strings.TrimSpace(s.cfg.DisplayName); displayName != "" {
		fromHeader = (&mail.Address{Name: displayName, Address: s.cfg.Address}).String()
	}

	msg, err := buildMessage(fromHeader, c.Value, feed, items, cfg.Format)
	if err != nil {
		return &worker.PermanentError{Msg: err.Error()}
	}

	return s.sendSMTP(ctx, c.Value, msg)
}

func (s *Sender) sendSMTP(_ context.Context, to string, msg []byte) error {
	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprintf("%d", s.cfg.SMTPPort))
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.cfg.SMTPHost, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(s.cfg.Address); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return &worker.PermanentError{Msg: fmt.Sprintf("invalid email recipient: %v", err)}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func buildMessage(from string, to string, feed domain.Feed, items []domain.Item, format string) ([]byte, error) {
	if _, err := mail.ParseAddress(to); err != nil {
		return nil, fmt.Errorf("invalid email recipient: %w", err)
	}

	subject := render.Template("{feed_name}: {title}", feed, items)
	if len(items) > 1 {
		subject = render.Template("{feed_name}: {item_count} new items", feed, items)
	}

	contentType := "text/plain; charset=UTF-8"
	body := render.Message(feed, items)
	if strings.EqualFold(strings.TrimSpace(format), "html") {
		contentType = "text/html; charset=UTF-8"
		body = render.HTMLMessage(feed, items)
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: " + contentType + "\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String()), nil
}
