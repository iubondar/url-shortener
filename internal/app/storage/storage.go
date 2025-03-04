package storage

import (
	"context"
	"errors"
)

// специальные типы ошибок
var ErrorNotFound = errors.New("not found")

type Repository interface {
	SaveURL(url string) (id string, exists bool, err error)
	RetrieveURL(id string) (url string, err error)
}

type StatusChecker interface {
	CheckStatus(ctx context.Context) error
}
