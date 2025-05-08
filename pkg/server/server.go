package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zhashkevych/go-pocket-sdk"
	"github.com/zhashkevych/telegram-pocket-bot/pkg/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AuthServer struct { // Определяем сервер аутентификации
	redirectUrl string // URL для редиректа

	storage storage.TokenStorage // Хранилище токенов
	client  *pocket.Client       // Клиент Pocket

	logger *zap.Logger
	server *http.Server
}

// Определяем новый сервер аутентификации. Пока что без логгера и сервера
func NewAuthServer(redirectUrl string, storage storage.TokenStorage, client *pocket.Client) *AuthServer {
	return &AuthServer{
		redirectUrl: redirectUrl,
		storage:     storage,
		client:      client,
	}
}

func (s *AuthServer) Start() error { // запуск Сервера аутентификации
	s.server = &http.Server{ // Сервер это HTTP сервер
		Handler: s, // <- Ключевой момент: AuthServer сам становится обработчиком!
		Addr:    ":80",
	}

	logger, _ := zap.NewDevelopment(zap.Fields( // Создаёт логгер для разработки (цветной вывод в консоль).
		zap.String("app", "authorization server"))) // добавляет статическое поле app: "authorization server" к каждому сообщению.
	defer logger.Sync() // Сбрасывает буфер логов перед выходом

	s.logger = logger
	s.logger.Info("Сервер запущен", zap.String("port", "80"))
	return s.server.ListenAndServe() //  На выходе запускаем цикл по серверу слушая запросы
}

func (s *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) { // Запускаем сервер
	if r.Method != http.MethodGet { // 1. Проверяет, что запрос использует метод GET.
		s.logger.Debug("received unavailable HTTP method request",
			zap.String("method", r.Method))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	chatIDQuery := r.URL.Query().Get("chat_id") // 2. Извлекает параметр chat_id из URL
	if chatIDQuery == "" {
		s.logger.Debug("received empty chat_id query param")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDQuery, 10, 64) // 3. Преобразует chat_id из строки в число (int64)
	if err != nil {
		s.logger.Debug("received invalid chat_id query param",
			zap.String("chat_id", chatIDQuery))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.createAccessToken(r.Context(), chatID); err != nil { // 4. Вызывает метод для генерации токена
		s.logger.Debug("failed to create access token",
			zap.String("err", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", s.redirectUrl)  // выполняет HTTP-редирект
	w.WriteHeader(http.StatusMovedPermanently) // Отправляет статус-код 301 (постоянное перенаправление)
}

func (s *AuthServer) createAccessToken(ctx context.Context, chatID int64) error { // Создание и сохранение токена доступа
	requestToken, err := s.storage.Get(chatID, storage.RequestTokens) // 1. Получение request-токена из хранилища
	if err != nil {
		return errors.WithMessage(err, "failed to get request token")
	}

	authResp, err := s.client.Authorize(ctx, requestToken) // 2. Авторизация в Pocket
	if err != nil {
		return errors.WithMessage(err, "failed to authorize at Pocket")
	}

	//3. Сохранение access-токена
	if err := s.storage.Save(chatID, authResp.AccessToken, storage.AccessTokens); err != nil {
		return errors.WithMessage(err, "failed to save access token to storage")
	}

	return nil
}
