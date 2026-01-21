package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/repository/postgres"
)

type Bot struct {
	api *tgbotapi.BotAPI
	db  *postgres.DB
	log *slog.Logger
}

func New(api *tgbotapi.BotAPI, db *postgres.DB, log *slog.Logger) *Bot {
	return &Bot{api: api, db: db, log: log}
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	b.log.Info("tg-bot polling started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case upd := <-updates:
			if upd.Message != nil {
				if err := b.onMessage(ctx, upd.Message); err != nil {
					b.log.Warn("onMessage failed", "err", err)
				}
				continue
			}
			if upd.CallbackQuery != nil {
				if err := b.onCallback(ctx, upd.CallbackQuery); err != nil {
					b.log.Warn("onCallback failed", "err", err)
				}
				continue
			}
		}
	}
}

func (b *Bot) onMessage(ctx context.Context, m *tgbotapi.Message) error {
	if m.Text == "" {
		return nil
	}
	if !strings.HasPrefix(m.Text, "/start") {
		return nil
	}

	msg := tgbotapi.NewMessage(m.Chat.ID,
		"Привет! Нажми кнопку ниже, чтобы подтвердить подписку на рассылку.")
	kb := confirmKeyboard()
	msg.ReplyMarkup = kb

	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) onCallback(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	if q.Data != cbConfirm {
		return nil
	}

	chatID := fmt.Sprintf("%d", q.Message.Chat.ID)

	var username *string
	if q.From != nil && q.From.UserName != "" {
		u := q.From.UserName
		username = &u
	}

	var displayName *string
	if q.From != nil {
		name := strings.TrimSpace(strings.Join([]string{q.From.FirstName, q.From.LastName}, " "))
		if name != "" {
			displayName = &name
		}
	}

	verifiedAt := time.Now().UTC()

	_, err := b.db.Contacts().UpsertTelegramActive(ctx, chatID, username, displayName, verifiedAt)
	if err != nil {
		_ = b.answerCallback(q.ID, "Ошибка подтверждения 😢")
		return err
	}

	_ = b.answerCallback(q.ID, "Готово! ✅")

	edit := tgbotapi.NewEditMessageText(q.Message.Chat.ID, q.Message.MessageID,
		"Подписка подтверждена ✅\nТеперь админ может добавить тебе нужные RSS-ленты.")
	_, _ = b.api.Send(edit)

	return nil
}

func (b *Bot) answerCallback(callbackID, text string) error {
	cfg := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Request(cfg)
	return err
}
