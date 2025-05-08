package boltdb

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/storage"
	"strconv"
)

type TokenStorage struct { // ввод структуры базы данных
	db *bolt.DB
}

func NewTokenStorage(db *bolt.DB) *TokenStorage {
	return &TokenStorage{db: db}
} // Конструктор

func (s *TokenStorage) Save(chatID int64, token string, bucket storage.Bucket) error { // Сохранение токена с чатайди
	return s.db.Update(func(tx *bolt.Tx) error { // Обновление в Болте
		b := tx.Bucket([]byte(bucket))                  // получение бакета по имени
		return b.Put(intToBytes(chatID), []byte(token)) // запись в бакет айдишника и токена
	})
}

func (s *TokenStorage) Get(chatID int64, bucket storage.Bucket) (string, error) { // Получить токен из бакета
	var token string

	err := s.db.View(func(tx *bolt.Tx) error { // Чтение данных из базы
		b := tx.Bucket([]byte(bucket))            // берем бакет с переданным именем
		token = string(b.Get(intToBytes(chatID))) // читаем значение по ключу байтового значения айдишника
		return nil
	})

	if token == "" { // если токен пустой, значит не найден
		return "", errors.New("not found")
	}

	return token, err // возвращаем токен и ошибку
}

func intToBytes(v int64) []byte {
	return []byte(strconv.FormatInt(v, 10))
} // переводим стрингконверсией инт в байты
