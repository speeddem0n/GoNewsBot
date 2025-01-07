package botkit

import (
	"context"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Bot struct { // Структура для тг бота
	api      *tgbotapi.BotAPI
	cmdViews map[string]ViewFunc // Мап для ViewFunc (В качестве кюча испольльзуется команда для бота)
}

// addsource (команда для добавления источников в бд)
// listsources (команда для получения списка источников)
// deletesource (команда для удаления) источника
type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error // Функция которая будет реагировать на определенную команду

/* tgbotapi.Update любой ивент который приходит от телеграма при взаимодействии с ботом
bot *tgbotapi.BotAPI клиет для доступа к боту */

func NewBot(api *tgbotapi.BotAPI) *Bot { // конструктор для структуры бота
	return &Bot{
		api: api,
	}
}

func (b *Bot) Start(ctx context.Context) error { // Метод для запуска бота
	u := tgbotapi.NewUpdate(0) // Устанавливаем канал в который будут писаться сообщения
	u.Timeout = 60             // Устанавка таймаута на 60 секунд

	updates := b.api.GetUpdatesChan(u) // Получаем сам chan

	for {
		select {
		case update := <-updates: // Кейс, когда получаем апдейт из канала updates
			updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Second) // Создаем новый контекст для обработки апдейта с таймаутом 5 секунд

			b.handleUpdate(updateCtx, update) // Вызываем метод handleUpdate
			updateCancel()                    // Отменяем контекст
		case <-ctx.Done(): // Кейс когда контекст завершен
			return ctx.Err()
		}

	}
}

func (b *Bot) RegisterCmdView(cmd string, view ViewFunc) { // Метод для регистрации View
	if b.cmdViews == nil { // Проверка, инициализирована ли мапа
		b.cmdViews = make(map[string]ViewFunc) // Если нет то инициализируем
	}

	b.cmdViews[cmd] = view // Добовляем команду в мапу
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) { // Метод для обработки tgbotapi.Update и направления их на соответствующие ViewFunc(комманды)
	defer func() { // В ViewFunc может произойти паника, отлавливаем ее с помощью recover() и логируем
		if p := recover(); p != nil {
			logrus.Errorf("panic recovered: %v\n%s", p, string(debug.Stack()))
		}
	}()

	var view ViewFunc // Переменная для ViewFunc

	if !update.Message.IsCommand() { // Если сообщение не содержит никакой команды то просто выходим
		return
	}

	cmd := update.Message.Command() // Достаем команду из сообщения

	cmdView, ok := b.cmdViews[cmd] // Пробуем достать View из мапы
	if !ok {
		return // Просто выходим если нет такой команды
	}

	view = cmdView

	if err := view(ctx, b.api, update); err != nil { // Вызываем view и обробатываем ошибку
		logrus.Errorf("failed to handle update: %v", err)

		if _, err := b.api.Send( // Отправляем пользователю сообщение об ошибке
			tgbotapi.NewMessage(update.Message.Chat.ID, "internal error"),
		); err != nil {
			logrus.Errorf("failed to send message: %v", err)
		}
	}
}
