package botkit

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Bot struct { // Структура для тг бота
	api *tgbotapi.BotAPI
}
