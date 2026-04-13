package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	cbConfirm = "confirm"
)

func confirmKeyboard(code string) tgbotapi.InlineKeyboardMarkup {
	data := cbConfirm
	if code != "" {
		data = cbConfirm + "|" + code
	}
	btn := tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить подписку", data)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
}
