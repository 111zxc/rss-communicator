package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/senders/render"
)

type Sender struct {
	api *tgbotapi.BotAPI
}

func New(api *tgbotapi.BotAPI) *Sender { return &Sender{api: api} }

func (s *Sender) Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	if len(items) == 0 {
		return nil
	}

	msgText := render.Message(feed, items)

	chatID, err := parseChatID(c.Value)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.DisableWebPagePreview = false

	_, err = s.api.Send(msg)
	if err != nil {
		return err
	}
	_ = ctx
	return nil
}

func parseChatID(v string) (int64, error) {
	var id int64
	_, err := fmt.Sscan(v, &id)
	return id, err
}
