package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zhashkevych/go-pocket-sdk"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/config"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/server"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/storage"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/storage/boltdb"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/telegram"
	"log"
)

func main() {
	cfg, err := config.Init() // Инициация конфигов
	if err != nil {
		log.Fatal(err) // если ошибка, фаталим
	}

	fmt.Println("App has initiated") // Вывод на экран

	botApi, err := tgbotapi.NewBotAPI(cfg.TelegramToken) // Запуск телеграм бота. С передачей токена
	if err != nil {
		log.Fatal(err) // если ошибка, фаталим
	}
	botApi.Debug = true //включает режим отладки

	pocketClient, err := pocket.NewClient(cfg.PocketConsumerKey) // запускаем клиент покета
	if err != nil {
		log.Fatal(err)
	}

	db, err := initBolt() // инициализация BoltDB
	if err != nil {
		log.Fatal(err) // если ошибка, фаталим
	}
	storage := boltdb.NewTokenStorage(db) //создание хранилища токенов

	bot := telegram.NewBot(botApi, pocketClient, cfg.AuthServerURL, storage, cfg.Messages) //создаёт экземпляр бота с данными:Основной клиент для работы с Telegram Bot API, Клиент для работы с API Pocket, URL сервера аутентификации, Хранилище токенов, Тексты сообщений бота

	redirectServer := server.NewAuthServer(cfg.BotURL, storage, pocketClient) // создаём HTTP-сервер для обработки OAuth-редиректов

	go func() {
		if err := redirectServer.Start(); err != nil {
			log.Fatal(err)
		} // Запуск сервера
	}()

	if err := bot.Start(); err != nil { // запуск бота
		log.Fatal(err)
	}
}

func initBolt() (*bolt.DB, error) { //  Инициализация Болта
	db, err := bolt.Open("bot.db", 0600, nil) // открываем болт
	if err != nil {
		return nil, err // выходим если ошибка
	}

	if err := db.Batch(func(tx *bolt.Tx) error {
		// 1. Создаём бакет для AccessTokens (если не существует)
		_, err := tx.CreateBucketIfNotExists([]byte(storage.AccessTokens))
		if err != nil {
			return err
		}
		
		// 2. Создаём бакет для RequestTokens (если не существует)
		_, err = tx.CreateBucketIfNotExists([]byte(storage.RequestTokens))
		return err
	}); err != nil {
		return nil, err
	}

	return db, nil
}
