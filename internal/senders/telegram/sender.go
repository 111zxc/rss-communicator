package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
)

type Sender struct {
	api *tgbotapi.BotAPI
}

func New(api *tgbotapi.BotAPI) *Sender { return &Sender{api: api} }

func (s *Sender) Send(ctx context.Context, c domain.Contact, feed domain.Feed, items []domain.Item) error {
	if len(items) == 0 {
		return nil
	}

	msgText := renderMessage(feed, items)

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

func renderMessage(feed domain.Feed, items []domain.Item) string {
	if len(items) == 1 {
		return fmt.Sprintf("📰 %s\n%s", items[0].Title, items[0].Link)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📰 %s\n%d new items\n\n", feed.Name, len(items))
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%d. %s\n%s", i+1, item.Title, item.Link)
	}
	return b.String()
}

func parseChatID(v string) (int64, error) {
	var id int64
	_, err := fmt.Sscan(v, &id)
	return id, err
}
