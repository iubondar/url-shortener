package models

import (
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
