// Package storage предоставляет интерфейсы и реализации для хранения и управления URL-ссылками.
// Поддерживает различные типы хранилищ: in-memory, файловое и PostgreSQL.
package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrorNotFound возвращается, когда запрашиваемая запись не найдена в хранилище.
var ErrorNotFound = errors.New("not found")

// Record представляет запись URL в хранилище.
type Record struct {
	ShortURL    string    `json:"short_url"`    // короткий идентификатор URL
	OriginalURL string    `json:"original_url"` // оригинальный URL
	UserID      uuid.UUID `json:"user_id"`      // идентификатор пользователя
	IsDeleted   bool      `json:"is_deleted"`   // флаг удаления
}

// Repository определяет интерфейс для работы с хранилищем URL.
// Предоставляет методы для сохранения, получения и удаления URL.
type Repository interface {
	// SaveURL сохраняет URL в хранилище.
	// Возвращает короткий идентификатор, флаг существования и ошибку.
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)

	// SaveURLs сохраняет массив URL в хранилище.
	// Возвращает массив коротких идентификаторов и ошибку.
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)

	// RetrieveByShortURL получает запись по короткому идентификатору.
	// Возвращает запись и ошибку.
	RetrieveByShortURL(ctx context.Context, shortURL string) (record Record, err error)

	// RetrieveUserURLs получает все URL пользователя.
	// Возвращает массив записей и ошибку.
	RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []Record, err error)

	// DeleteByShortURLs помечает URL как удаленные.
	// Принимает идентификатор пользователя и массив коротких идентификаторов.
	DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string)

	// StatusChecker предоставляет методы для проверки состояния хранилища.
	StatusChecker
}
