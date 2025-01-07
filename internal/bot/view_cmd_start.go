package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
)

func ViewCmdStart() botkit.ViewFunc { // View для запуска бота
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if _, err := bot.Send(tgbotapi.NewMessage(update.FromChat().ID, // Отправляем пользователь сообщение о запуске
			"Hello, im news bot, i can post news in your telegram channel")); err != nil {
			return err
		}

		return nil
	}
}
