package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// специальные типы ошибок
var ErrorNotFound = errors.New("not found")

type Record struct {
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      uuid.UUID `json:"user_id"`
	IsDeleted   bool      `json:"is_deleted"`
}

type Repository interface {
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)
	RetrieveByShortURL(ctx context.Context, shortURL string) (record Record, err error)
	RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []Record, err error)
	StatusChecker
}
