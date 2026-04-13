package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/111zxc/rss-communicator/internal/repository/postgres"
	"github.com/111zxc/rss-communicator/internal/service"
)

type Bot struct {
	api *tgbotapi.BotAPI
	db  *postgres.DB
	reg *service.RegistrationService
	log *slog.Logger
}

func New(api *tgbotapi.BotAPI, db *postgres.DB, reg *service.RegistrationService, log *slog.Logger) *Bot {
	return &Bot{api: api, db: db, reg: reg, log: log}
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
	cmd, code, ok := parseRegistrationCommand(m.Text)
	if !ok {
		return nil
	}

	text := "Привет! Нажми кнопку ниже, чтобы подтвердить регистрацию."
	if code != "" {
		text = fmt.Sprintf("Привет! Нажми кнопку ниже, чтобы подтвердить регистрацию по коду `%s`.", code)
	}
	if cmd == "register" && code == "" {
		text = "Привет! Нажми кнопку ниже, чтобы подтвердить регистрацию."
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, text)
	msg.ParseMode = "Markdown"
	kb := confirmKeyboard(code)
	msg.ReplyMarkup = kb

	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) onCallback(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	code, ok := parseConfirmCallback(q.Data)
	if !ok {
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

	result, err := b.reg.RegisterTelegram(ctx, service.RegisterTelegramInput{
		ChatID:      chatID,
		Username:    username,
		DisplayName: displayName,
		Code:        code,
	})
	if err != nil {
		if err == service.ErrAlreadyRegistered {
			_ = b.answerCallback(q.ID, "Ты уже зарегистрирован")
			return nil
		}
		switch err {
		case service.ErrRegistrationCodeNotFound:
			_ = b.answerCallback(q.ID, "Код регистрации не найден")
			return nil
		case service.ErrRegistrationCodeDisabled:
			_ = b.answerCallback(q.ID, "Код регистрации отключен")
			return nil
		case service.ErrRegistrationCodeExpired:
			_ = b.answerCallback(q.ID, "Срок действия кода истек")
			return nil
		case service.ErrRegistrationCodeExhausted:
			_ = b.answerCallback(q.ID, "Лимит использований кода исчерпан")
			return nil
		}
		_ = b.answerCallback(q.ID, "Ошибка подтверждения 😢")
		return err
	}

	_ = b.answerCallback(q.ID, "Готово! ✅")

	text := "Регистрация подтверждена ✅"
	if result.AppliedCode != nil && len(result.AppliedGroups) > 0 {
		names := make([]string, 0, len(result.AppliedGroups))
		for _, group := range result.AppliedGroups {
			names = append(names, group.Name)
		}
		text = fmt.Sprintf("Регистрация подтверждена ✅\nКод `%s` применён.\nДобавлен(а) в группы: %s.", result.AppliedCode.Code, strings.Join(names, ", "))
	} else if result.AppliedCode != nil {
		text = fmt.Sprintf("Регистрация подтверждена ✅\nКод `%s` применён.", result.AppliedCode.Code)
	}

	edit := tgbotapi.NewEditMessageText(q.Message.Chat.ID, q.Message.MessageID, text)
	edit.ParseMode = "Markdown"
	_, _ = b.api.Send(edit)

	return nil
}

func (b *Bot) answerCallback(callbackID, text string) error {
	cfg := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Request(cfg)
	return err
}

func parseRegistrationCommand(text string) (string, string, bool) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return "", "", false
	}
	switch parts[0] {
	case "/start", "/start@rss_communicator_bot":
		if len(parts) > 1 {
			return "start", strings.ToUpper(strings.TrimSpace(parts[1])), true
		}
		return "start", "", true
	case "/register", "/register@rss_communicator_bot":
		if len(parts) > 1 {
			return "register", strings.ToUpper(strings.TrimSpace(parts[1])), true
		}
		return "register", "", true
	default:
		return "", "", false
	}
}

func parseConfirmCallback(data string) (string, bool) {
	if data == cbConfirm {
		return "", true
	}
	prefix := cbConfirm + "|"
	if strings.HasPrefix(data, prefix) {
		return strings.TrimPrefix(data, prefix), true
	}
	return "", false
}
