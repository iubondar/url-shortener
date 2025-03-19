package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// специальные типы ошибок
var ErrorNotFound = errors.New("not found")

type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Repository interface {
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)
	RetrieveURL(ctx context.Context, id string) (url string, err error)
	RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (URLPairs []URLPair, err error)
	StatusChecker
}
