package senders

import (
	"context"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type Sender interface {
	Send(ctx context.Context, c domain.Contact, item domain.Item) error
}
