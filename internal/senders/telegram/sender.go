package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type Sender struct {
	api *tgbotapi.BotAPI
}

func New(api *tgbotapi.BotAPI) *Sender { return &Sender{api: api} }

func (s *Sender) Send(ctx context.Context, c domain.Contact, item domain.Item) error {
	msgText := fmt.Sprintf("📰 %s\n%s", item.Title, item.Link)

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
