package config

import (
	"github.com/spf13/viper"
	"os"
)

type Messages struct { // Структура сообщений
	Responses
	Errors
}

type Responses struct { // Структура Ответов
	Start             string `mapstructure:"start"`
	AlreadyAuthorized string `mapstructure:"already_authorized"`
	UnknownCommand    string `mapstructure:"unknown_command"`
	LinkSaved         string `mapstructure:"link_saved"`
}

type Errors struct { // Структура сообщений об ошибках
	Default      string `mapstructure:"default"`
	InvalidURL   string `mapstructure:"invalid_url"`
	UnableToSave string `mapstructure:"unable_to_save"`
}

type Config struct { // Структура конфигов
	TelegramToken     string // берется из .env
	PocketConsumerKey string // берется из .env
	AuthServerURL     string // берется из .env

	BotURL     string `mapstructure:"bot_url"`
	BoltDBFile string `mapstructure:"db_file"`

	Messages Messages // Привязываются тексты сообщений
}

func Init() (*Config, error) {
	if err := setUpViper(); err != nil { // Запускаем вайпер
		return nil, err // выходим если ошибка
	}

	var cfg Config // объявляем переменную конфигов
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	} // анмаршалим конфиги из .yml

	if err := fromEnv(&cfg); err != nil {
		return nil, err
	} // достаем конфиги из .env

	return &cfg, nil
}

func unmarshal(cfg *Config) error { //
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("messages.response", &cfg.Messages.Responses); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("messages.error", &cfg.Messages.Errors); err != nil {
		return err
	}

	return nil
}

//TOKEN=7677898491:AAF9oEy831OpShl61sYZw0nT_s2e6ZRb5NA
//CONSUMER_KEY=114575-2848d9b18c1bad00c4283da
//AUTH_SERVER_URL=http://localhost/

func fromEnv(cfg *Config) error {
	os.Setenv("TOKEN", "7677898491:AAF9oEy831OpShl61sYZw0nT_s2e6ZRb5NA")
	os.Setenv("CONSUMER_KEY", "114575-2848d9b18c1bad00c4283da")
	os.Setenv("AUTH_SERVER_URL", "http://localhost/")

	if err := viper.BindEnv("token"); err != nil {
		return err
	} // 1. Сначала привязываем переменную окружения к Viper
	cfg.TelegramToken = viper.GetString("token") // 2. Затем получаем значение

	if err := viper.BindEnv("consumer_key"); err != nil {
		return err
	} // привязываем
	cfg.PocketConsumerKey = viper.GetString("consumer_key") // получаем значение

	if err := viper.BindEnv("auth_server_url"); err != nil {
		return err
	} // привязываем
	cfg.AuthServerURL = viper.GetString("auth_server_url") // получаем значение

	return nil
}

func setUpViper() error {
	viper.AddConfigPath("configs") // путь папка с main.yml
	viper.SetConfigName("main")    // имя файла main.yml

	return viper.ReadInConfig()
}
