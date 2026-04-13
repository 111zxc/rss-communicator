package senders

import (
	"context"
	"fmt"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
)

type Router struct {
	telegram Sender
	email    Sender
	http     Sender
}

func NewRouter(telegram Sender, email Sender, http Sender) *Router {
	return &Router{telegram: telegram, email: email, http: http}
}

func (r *Router) Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	switch c.Type {
	case domain.ContactTelegram:
		if r.telegram == nil {
			return &worker.PermanentError{Msg: "telegram sender is not configured"}
		}
		return r.telegram.Send(ctx, c, feed, items)
	case domain.ContactEmail:
		if r.email == nil {
			return &worker.PermanentError{Msg: "email sender is not configured"}
		}
		return r.email.Send(ctx, c, feed, items)
	case domain.ContactHTTP:
		if r.http == nil {
			return &worker.PermanentError{Msg: "http sender is not configured"}
		}
		return r.http.Send(ctx, c, feed, items)
	default:
		return &worker.PermanentError{Msg: fmt.Sprintf("unsupported contact type: %s", c.Type)}
	}
}
