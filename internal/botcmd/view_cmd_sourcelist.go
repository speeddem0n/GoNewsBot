package botcmd

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
	"github.com/speeddem0n/GoNewsBot/internal/botkit/markup"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type SourceLister interface { // Интерфейс для работы со слоем storage
	Sources(ctx context.Context) ([]models.Source, error)
}

func ViewCmdListSources(lister SourceLister) botkit.ViewFunc { // View для вывода списка всех источников
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		sources, err := lister.Sources(ctx)
		if err != nil {
			return err
		}

		var (
			sourcesInfo = lo.Map(sources, func(source models.Source, _ int) string {
				return formatSource(source)
			}) // Форматируем список источников в читаймый вид
			msgText = fmt.Sprintf("Список источников\\(Всего %d\\):\n\n%s", len(sources), strings.Join(sourcesInfo, "\n\n")) // Финальное сообзение для пользователя
		)

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, msgText) // Формеруем ответ пользователю
		reply.ParseMode = "MarkdownV2"                                // Ответ в формате MarkdownV2

		if _, err := bot.Send(reply); err != nil { // Отправляем сообщение пользователю
			return err
		}

		return nil
	}
}

func formatSource(source models.Source) string { // Функция для форматирования инфо об источнике
	return fmt.Sprintf("🌎 *%s*\nID: `%d`\nURL feed: %s",
		markup.EscapeForMarkdown(source.Name), // Функция для замены спецсимволов markdown в тексте
		source.ID,
		markup.EscapeForMarkdown(source.FeedURL), // Функция для замены спецсимволов markdown в тексте
	)
}
