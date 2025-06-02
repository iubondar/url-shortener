package handlers

import (
	"context"

	"github.com/google/uuid"
)

// URLSaver определяет интерфейс для сохранения URL в хранилище.
type URLSaver interface {
	// SaveURL сохраняет URL в хранилище.
	// Возвращает короткий идентификатор, флаг существования и ошибку.
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
}
