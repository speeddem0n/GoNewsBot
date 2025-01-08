package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	bot "github.com/speeddem0n/GoNewsBot/internal/botcmd"
	"github.com/speeddem0n/GoNewsBot/internal/botkit"
	"github.com/speeddem0n/GoNewsBot/internal/config"
	"github.com/speeddem0n/GoNewsBot/internal/fetcher"
	"github.com/speeddem0n/GoNewsBot/internal/notifier"
	"github.com/speeddem0n/GoNewsBot/internal/storage"
	"github.com/speeddem0n/GoNewsBot/internal/summary"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken) // Создадин новый tgbotAPI
	if err != nil {
		logrus.Errorf("failed to create bot: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN) // Подключаемся к бд
	if err != nil {
		logrus.Errorf("failed to connect to database: %v", err)
		return
	}
	defer db.Close() // Откладываем закрытие соеденения с бд

	var ( // Инициализация зависимостей
		articleStorage = storage.NewArticleStorage(db) // Слой хранилища статей
		sourceStorage  = storage.NewSourceStorage(db)  // Слой хранилища источников
		fetcher        = fetcher.NewFetcher(           // Слой fetcher который забирает статьи из источников
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		notifier = notifier.NewNotifier( // слой notifier
			articleStorage,
			summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			botAPI,
			config.Get().NotificationInterval,
			config.Get().LookupTimeWindow, // lookupTimeWindow равен двум FetchInterval
			config.Get().TelegramChannelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // Контекст для Graceful shutdown
	defer cancel()

	newsBot := botkit.NewBot(botAPI)                                       // Инициализируем тг бота
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())                   // Инициализируем View для команды start
	newsBot.RegisterCmdView("add", bot.ViewCmdAddSource(sourceStorage))    // Инициализируем View для команды add
	newsBot.RegisterCmdView("list", bot.ViewCmdListSources(sourceStorage)) // Инициализируем View для команды list

	go func(ctx context.Context) { // Запуск первого воркера (Fetcher)
		if err := fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) { // если ошибка != остановке контекста, логируем и выходим из горутины
				logrus.Errorf("failed to start fetcher: %v", err)
				return
			}

			logrus.Println("fetcher stopped")
		}

	}(ctx)

	go func(ctx context.Context) { // Запуск второго воркера Notifier
		if err := notifier.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) { // если ошибка != остановке контекста, логируем и выходим из горутины
				logrus.Errorf("failed to start notifier: %v", err)
				return
			}

			logrus.Println("notifier stopped")
		}
	}(ctx)

	if err := newsBot.Start(ctx); err != nil { // Запуск телеграм бота
		if !errors.Is(err, context.Canceled) { // если ошибка != остановке контекста, логируем и выходим из горутины
			logrus.Errorf("failed to start tg bot: %v", err)
			return
		}

		logrus.Println("bot stopped")
	}
}
