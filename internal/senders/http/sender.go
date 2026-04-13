package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
	"github.com/111zxc/rss-communicator/internal/senders/render"
)

type ContactsRepository interface {
	GetHTTPConfig(ctx context.Context, contactID string) (domain.HTTPContactConfig, error)
}

type Sender struct {
	contacts ContactsRepository
	client   *http.Client
}

func New(contacts ContactsRepository, client *http.Client) *Sender {
	if client == nil {
		client = http.DefaultClient
	}
	return &Sender{contacts: contacts, client: client}
}

func (s *Sender) Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	cfg := c.HTTP
	if cfg == nil {
		loaded, err := s.contacts.GetHTTPConfig(ctx, c.ID)
		if err != nil {
			return &worker.PermanentError{Msg: fmt.Sprintf("http config not found: %v", err)}
		}
		cfg = &loaded
	}

	method := strings.ToUpper(strings.TrimSpace(cfg.Method))
	if method == "" {
		method = http.MethodPost
	}

	targetURL := render.Template(cfg.URL, feed, items)

	var body io.Reader
	if cfg.BodyTemplate != nil {
		rendered := render.Template(*cfg.BodyTemplate, feed, items)
		body = strings.NewReader(rendered)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return &worker.PermanentError{Msg: fmt.Sprintf("invalid http request: %v", err)}
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, render.Template(v, feed, items))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	msg := fmt.Sprintf("http status %d", resp.StatusCode)
	if snippet, err := io.ReadAll(io.LimitReader(resp.Body, 2048)); err == nil {
		text := strings.TrimSpace(string(snippet))
		if text != "" {
			msg = fmt.Sprintf("%s: %s", msg, text)
		}
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return &worker.PermanentError{Msg: msg}
	}
	return fmt.Errorf("%s", msg)
}
