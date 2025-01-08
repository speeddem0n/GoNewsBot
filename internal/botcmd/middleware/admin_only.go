package middleware

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
)

func AdminOnly(channelID int64, next botkit.ViewFunc) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		admins, err := bot.GetChatAdministrators( // Получаем список админов канала
			tgbotapi.ChatAdministratorsConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: channelID,
				},
			},
		)
		if err != nil {
			return err
		}

		for _, admin := range admins { // Проходимся по списку админов
			if admin.User.ID == update.Message.From.ID { // Проверяем есть ли ID пользователя в ID администраторов
				return next(ctx, bot, update)
			}
		}

		if _, err := bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"У вас нет прав для выполнения данной команды",
		)); err != nil {
			return err
		}

		return nil
	}
}
