package storage

type Bucket string

const (
	AccessTokens  Bucket = "access_tokens"  // Название бакета с токенами доступа
	RequestTokens Bucket = "request_tokens" // Название бакета с токенами запроса
)

type TokenStorage interface { // Определяет контракт
	Save(chatID int64, token string, bucket Bucket) error
	Get(chatID int64, bucket Bucket) (string, error)
}
