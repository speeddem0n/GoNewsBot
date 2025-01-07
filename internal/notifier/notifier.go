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

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type Notifier struct {
	articles         ArticleProvider
	summarizer       Summarizer
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
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

func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	topeOneArticles, err := n.articles.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow), 1)
	if err != nil {
		logrus.Printf("Error on getting not posted article: %s", err)
		return err
	}

	if len(topeOneArticles) == 0 {
		logrus.Printf("All articles are posted")
		return nil
	}

	article := topeOneArticles[0]

	summary, err := n.extractSummary(ctx, article)
	if err != nil {
		logrus.Printf("Error on extract summary: %s", err)
		return err
	}

	if err := n.sendArticle(article, summary); err != nil {
		logrus.Printf("Error on send article: %s", err)
		return err
	}

	return n.articles.MarkPosted(ctx, article.ID)
}

func (n *Notifier) extractSummary(ctx context.Context, article models.Article) (string, error) { // Метод для получения Summary статьи
	var r io.Reader

	if article.Summary != "" { // Если у статьи есть Summary
		r = strings.NewReader(article.Summary) // Ридером будет Summary
	} else {
		resp, err := http.Get(article.Link) // Если у статьи нет Summary и переходем по ее адресу и забираем http body
		if err != nil {
			logrus.Printf("Error %s on request on %s", err, article.Link)
			return "", err
		}
		defer resp.Body.Close() // Откладываем закрытия тела ответа

		r = resp.Body // Ридером будет тело ответа
	}

	doc, err := readability.FromReader(r, nil) // Форматируем с помошью библеотеки readability html разметку страницы в читаймый документ
	if err != nil {
		logrus.Printf("Failed to parses an `io.Reader`: %s", err)
		return "", nil
	}

	summary, err := n.summarizer.Summarize(ctx, cleanText(doc.TextContent))
	if err != nil {
		logrus.Printf("Failed to get summary from summarizer.Summarize: %s", err)
		return "", err
	}

	return "\n\n" + summary, nil // Две пустые строки для отступа после заголовка
}

var redundantNewLines = regexp.MustCompile(`\n{3,}`) // Регулярка соответствует всем последовательностям пустых строк, где они идут 3 и более раз подряд

func cleanText(text string) string { // Функция для очистки текста от пустых строк
	return redundantNewLines.ReplaceAllString(text, "\n")
}

func (n *Notifier) sendArticle(article models.Article, summary string) error {
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
		logrus.Printf("Faildes to send msg to telegram: %s", err)
		return err
	}
}
