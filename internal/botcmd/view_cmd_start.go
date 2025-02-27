package botcmd

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
)

func ViewCmdStart() botkit.ViewFunc { // View для запуска бота
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if _, err := bot.Send(tgbotapi.NewMessage(update.FromChat().ID, // Отправляем пользователь сообщение о запуске
			botkit.CommandList)); err != nil {
			return err
		}

		return nil
	}
}
