package telegram

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zhashkevych/go-pocket-sdk"
	"net/url"
)

const (
	commandStart = "start"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error { // Обработчик событий
	switch message.Command() {
	case commandStart:
		return b.handleStartCommand(message) // обработчик команды старт
	default:
		return b.handleUnknownCommand(message) // Все остальные команды
	}
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) error {
	_, err := b.getAccessToken(message.Chat.ID) // проверяем наличие Аксесс токена
	if err != nil {
		return b.initAuthorizationProcess(message)
	} // если ошибка, авторизуем юзера

	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Responses.AlreadyAuthorized) // формируем сообщение с сообщением, что уже авторизован.
	_, err = b.bot.Send(msg)                                                            // отправляем сообщение
	return err
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Responses.UnknownCommand) // формируем сообщение с сообщением, что команда неизвестна
	_, err := b.bot.Send(msg)
	return err
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	accessToken, err := b.getAccessToken(message.Chat.ID) // получаем Аксесс токен
	if err != nil {
		return b.initAuthorizationProcess(message)
	} // если ошибка, авторизуем юзера

	if err := b.saveLink(message, accessToken); err != nil {
		return err
	} // сохраняем ссылку в Покет

	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Responses.LinkSaved) // формируем смс, что ссылка сохранена
	_, err = b.bot.Send(msg)                                                    // отправляем смс
	return err
}

func (b *Bot) saveLink(message *tgbotapi.Message, accessToken string) error {
	if err := b.validateURL(message.Text); err != nil {
		return invalidUrlError
	} // проверяем формат ссылки. ошибка невалидный урл

	if err := b.client.Add(context.Background(), pocket.AddInput{
		URL:         message.Text,
		AccessToken: accessToken,
	}); err != nil {
		return unableToSaveError
	} // добавляем ссылку к юзеру в покет

	return nil
}

func (b *Bot) validateURL(text string) error { // Проверяем смс на ссылку
	_, err := url.ParseRequestURI(text)
	return err
}
