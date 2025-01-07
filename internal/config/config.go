package config

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
	"github.com/sirupsen/logrus"
)

type Config struct { // Структура для конфига
	TelegramBotToken     string        `hcl:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN" required:"true"`
	TelegramChannelID    int64         `hcl:"telegram_channel_id" env:"TELEGRAM_CHANNEL_ID" required:"true"`
	DatabaseDSN          string        `hcl:"database_dsn" env:"DATABASE_DSN" default:"postgres://postgres:postgres@localhost:5438/news_feed_bot?sslmode=disable"`
	FetchInterval        time.Duration `hcl:"fetch_interval" env:"FETCH_INTERVAL" default:"10m"`
	NotificationInterval time.Duration `hcl:"notification_interval" env:"NOTIFICATION_INTERVAL" default:"1m"`
	FilterKeywords       []string      `hcl:"filter_keywords" env:"FILTER_KEYWORDS"`
	OpenAIKey            string        `hcl:"openai_key" env:"OPENAI_KEY"`
	OpenAIPrompt         string        `hcl:"openai_prompt" env:"OPENAI_PROMPT"`
}

var ( // Переменные cfg  для записи конфига и once sync.Once для выполнения операции только один раз
	cfg  Config
	once sync.Once
)

func Get() Config { // Функция для получения конфига
	once.Do(func() { // Получаем конфиг с помошью once.Do
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{ // LoaderFor создает новый Loader на основе структуры Config
			EnvPrefix: "NFB",                                                                                                                    // Задаем префиксс для переменных окружения
			Files:     []string{"./internal/config/config.hcl", "./internal/config/config.local.hcl", "$HOME/.config/news-feed-bot/config.hcl"}, // Файлы где хранится конфиг
			FileDecoders: map[string]aconfig.FileDecoder{ // Декодер для файлов hcl
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil { // Загружает конфигурацию
			logrus.Errorf("Error on loading config %s\n%s", err, string(debug.Stack())) // Логируем ошибку
		}
	})

	return cfg // Возвращаем структуру Config
}
