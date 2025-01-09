# Go News Bot
Telegram Bot для получения новостей из rss ленты сайтов и публикации их в тг канал.

# Что Умеет Бот
- Доставать новостные статьи из rss фида и публиковать их в тг канал
- Опционально делать запросы к ChatGPT для получения краткой выжимки из статьи
- Бот управляется с помощью админ команд

# Переменные окружения
- `NFB_TELEGRAM_BOT_TOKEN` — Токен для Telegram Bot API (Обязательный параметр)
- `NFB_TELEGRAM_CHANNEL_ID` — ID тг канала для публикации, можно узнать с помощью[@JsonDumpBot](https://t.me/JsonDumpBot)(Обязательный параметр)
- `NFB_DATABASE_DSN` — Строка для подключения к PostgreSQL
- `NFB_FETCH_INTERVAL` — Интервал для получения новых статей из источников, по умолчанию: 10 минут
- `NFB_NOTIFICATION_INTERVAL` — Интервал для публикации статьи в тг канал, по умолчанию: 1 минута
- `NFB_LOOKUP_TIME_WINDOW` — Максимальный срок давности публикуемой статьи
- `NFB_FILTER_KEYWORDS` — Список фильтрующих слов для пропуска ненужных статей
- `NFB_OPENAI_KEY` — токен для OpenAI API
- `NFB_OPENAI_PROMPT` — Текст запроса для GPT-3.5 Turbo что бы сгенерировать выжимку.

## HCL

Go News Bot может настраиваться с помощью HCL config файла. Сервис ищет config файлы по следующим путям:

- `./internal/config/config.hcl`
- `./internal/config/config.local.hcl`
- `$HOME/.config/news-feed-bot/config.hcl`

Имена параметров те же что и у переменных окружения, за исключением приставки NFB. Параметры записываются в нижнем регистре.
Пример: telegram_bot_token
