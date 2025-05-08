package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zhashkevych/go-pocket-sdk"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/config"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/storage"
)

type Bot struct { // Структура с ботом
	bot         *tgbotapi.BotAPI
	client      *pocket.Client
	redirectURL string

	storage storage.TokenStorage

	messages config.Messages
}

// Конструктор на Бота
func NewBot(bot *tgbotapi.BotAPI, client *pocket.Client, redirectURL string, storage storage.TokenStorage, messages config.Messages) *Bot {
	return &Bot{
		bot:         bot,
		client:      client,
		redirectURL: redirectURL,
		storage:     storage,
		messages:    messages,
	}
}

func (b *Bot) Start() error { // Начало работы бота
	u := tgbotapi.NewUpdate(0) // Создаёт объект для получения обновлений с offset = 0 (будут приходить только новые сообщения после запуска бота)
	u.Timeout = 60             // Устанавливает таймаут 60 секунд для long-polling запросов к Telegram API

	updates, err := b.bot.GetUpdatesChan(u) // Получает канал (chan tgbotapi.Update) с обновлениями от Telegram
	if err != nil {
		return err
	}

	for update := range updates { // Бесконечный цикл, который читает обновления из канала
		if update.Message == nil { // ignore any non-Message Updates. Пропускает обновления, не содержащие сообщений (например, нажатия кнопок, редактирование сообщений)
			continue
		}

		// Handle commands
		if update.Message.IsCommand() { //Если сообщение — команда (начинается с /), вызывает handleCommand()
			if err := b.handleCommand(update.Message); err != nil {
				b.handleError(update.Message.Chat.ID, err)
			}

			continue
		}

		// Handle regular messages. Обрабатывает текст, не являющийся командой (например, ссылки для сохранения в Pocket). При ошибке также уведомляет пользователя
		if err := b.handleMessage(update.Message); err != nil {
			b.handleError(update.Message.Chat.ID, err)
		}
	}

	return nil
}
