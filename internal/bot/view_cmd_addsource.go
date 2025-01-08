package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type SourceStorage interface {
	Add(ctx context.Context, source models.Source) (int64, error)
}

func ViewCmdAddSource(storage SourceStorage) botkit.ViewFunc { // View для добавления источника
	type addSourceArgs struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[addSourceArgs](update.Message.CommandArguments()) // парсим JSON объект из аргументов комманды в тип ddSourceArgs
		if err != nil {
			errReply := tgbotapi.NewMessage(update.Message.Chat.ID, invalidCommandMsg)
			errReply.ParseMode = "MarkdownV2"
			if _, err := bot.Send(errReply); err != nil {
				return err
			}
			return err
		}

		source := models.Source{ // Заполняем модель источника данными из args
			Name:    args.Name,
			FeedURL: args.URL,
		}

		sourceID, err := storage.Add(ctx, source)
		if err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf("Источник добавлен с ID: `%d`.\\ Используйте этот ID для управления источником\\.", sourceID) // Сообщение для пользователя
			reply   = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}

}
