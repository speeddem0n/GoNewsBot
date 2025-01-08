package botcmd

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
	"github.com/speeddem0n/GoNewsBot/internal/botkit/markup"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type SourceDeleter interface {
	Delete(ctx context.Context, id int64) error
}

func ViewCmdDelete(deleter SourceDeleter) botkit.ViewFunc {
	type deleteSourceArgs struct {
		ID int64 `json:"id"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[deleteSourceArgs](update.Message.CommandArguments())
		if err != nil {
			errReply := tgbotapi.NewMessage(update.Message.Chat.ID, markup.EscapeForMarkdown(botkit.InvalidDeleteInput))
			errReply.ParseMode = "MarkdownV2"
			if _, err := bot.Send(errReply); err != nil {
				return err
			}
			return err
		}

		source := models.Source{
			ID: args.ID,
		}

		if err = deleter.Delete(ctx, source.ID); err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf("Источник удален с ID: `%d`\\.", source.ID) // Сообщение для пользователя
			reply   = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "MarkdownV2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
