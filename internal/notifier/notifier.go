package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/speeddem0n/GoNewsBot/internal/botkit/markup"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type ArticleProvider interface { // Интейвейс для работы со стоем storage/article.go
	AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]models.Article, error) // Метод для получения всех неопубликованных статей
	MarkPosted(ctx context.Context, id int64) error                                            // Метод для отметки статьи как опубликованная
}

type Summarizer interface { // Интерфейс для связи со слоем openAPI
	Summarize(ctx context.Context, text string) (string, error)
}

type Notifier struct { // Структура notifier
	articles         ArticleProvider  // Интервейс для связи со слоем storage
	summarizer       Summarizer       // Интерфейс для связи со слоем openAPI
	bot              *tgbotapi.BotAPI // API tg бота
	sendInterval     time.Duration    // Интервал с которым бот публикует сообщения в канал
	lookupTimeWindow time.Duration    // Ограничение про времени публикации статьи которую бот будет постить
	channelID        int64            // Id канала
}

func NewNotifier(articleProvider ArticleProvider,
	summarizer Summarizer,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	channelID int64,
) *Notifier { // Конструктор для структуры Notifier
	return &Notifier{
		articles:         articleProvider,
		summarizer:       summarizer,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		channelID:        channelID,
	}
}

func (n *Notifier) Start(ctx context.Context) error { // Метод для запуска Notifier'a в отдельной горутине
	ticker := time.NewTicker(n.sendInterval) // Создаем тикер с заданым интервалом
	defer ticker.Stop()                      // Откладываем завершение тикера

	if err := n.SelectAndSendArticle(ctx); err != nil { // Первый SelectAndSendArticle запуск без ожидания интервала
		return err
	}

	for { // Бесконечный цикл
		select {
		case <-ticker.C: // Сработал тикер
			if err := n.SelectAndSendArticle(ctx); err != nil { // Вызываем SelectAndSendArticle
				return err
			}
		case <-ctx.Done(): // Контекст завершен
			return ctx.Err() // Возвращаем ошибку контекста
		}
	}
}

func (n *Notifier) SelectAndSendArticle(ctx context.Context) error { // Метод для выбора и отправки статьи
	topeOneArticles, err := n.articles.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow), 1) // Методом AllNotPosted достаем одну не опубликованную статью
	if err != nil {
		logrus.Errorf("Error on getting not posted article: %s", err)
		return err
	}

	if len(topeOneArticles) == 0 { // Проверяем есть вообще неопубликованная статья
		logrus.Info("All articles are posted")
		return nil
	}

	article := topeOneArticles[0] // Берем первую статью в переменную article

	summary, err := n.extractSummary(ctx, article) // получаем Summary статьи методом extractSummary
	if err != nil {
		logrus.Errorf("Error on extract summary: %s", err)
		return err
	}

	if err := n.sendArticle(article, summary); err != nil { // методом sendArticle публикуем статью в тг канал
		logrus.Errorf("Error on send article: %s", err)
		return err
	}

	return n.articles.MarkPosted(ctx, article.ID) // в конце вызываем метод MarkPosted и помечаем статью как опубликованную и возвращаем ошибку
}

func (n *Notifier) extractSummary(ctx context.Context, article models.Article) (string, error) { // Метод для получения Summary статьи
	var r io.Reader // Создаем новый объект io.Reader

	if article.Summary != "" { // Если у статьи есть Summary
		r = strings.NewReader(article.Summary) // Ридером будет Summary
	} else {
		resp, err := http.Get(article.Link) // Если у статьи нет Summary и переходем по ее адресу и забираем http body
		if err != nil {
			logrus.Errorf("Error %s on request on %s", err, article.Link)
			return "", err
		}
		defer resp.Body.Close() // Откладываем закрытия тела ответа

		r = resp.Body // Ридером будет тело ответа
	}

	doc, err := readability.FromReader(r, nil) // Форматируем с помошью библеотеки readability html разметку страницы в читаймый документ
	if err != nil {
		logrus.Errorf("Failed to parse an `io.Reader`: %s", err)
		return "", nil
	}

	summary, err := n.summarizer.Summarize(ctx, cleanText(doc.TextContent)) // Получаем summary методом Summarize
	if err != nil {
		logrus.Errorf("Failed to get summary from summarizer.Summarize: %s", err)
		return "", err
	}

	return "\n\n" + summary, nil // Две пустые строки для отступа после заголовка
}

var redundantNewLines = regexp.MustCompile(`\n{3,}`) // Регулярка соответствует всем последовательностям пустых строк, где они идут 3 и более раз подряд

func cleanText(text string) string { // Функция для очистки текста от пустых строк
	return redundantNewLines.ReplaceAllString(text, "\n")
}

func (n *Notifier) sendArticle(article models.Article, summary string) error { // Метод для публикации статьи в чат
	const msgFormat = "*%s*%s\n\n%s" // Шаблон сообщения

	msg := tgbotapi.NewMessage(n.channelID, fmt.Sprintf(
		msgFormat,
		markup.EscapeForMarkdown(article.Title), // Вызывается EscapeForMarkdown для замены Markdown спец символов
		markup.EscapeForMarkdown(summary),
		markup.EscapeForMarkdown(article.Link),
	)) // Создаем новое сообщение для бота

	msg.ParseMode = tgbotapi.ModeMarkdownV2 // Сообщение парсится как MarkdownV2 сообщение

	_, err := n.bot.Send(msg) // Send will send a Chattable item to Telegram.
	if err != nil {
		logrus.Errorf("Faildes to send msg to telegram: %s", err)
		return err
	}

	return nil
}
