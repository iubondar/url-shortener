// Package storage предоставляет интерфейсы и реализации для хранения и управления URL-ссылками.
package storage

import "context"

// StatusChecker определяет интерфейс для проверки состояния хранилища.
// Используется для проверки доступности и работоспособности хранилища.
type StatusChecker interface {
	// CheckStatus проверяет состояние хранилища.
	// Возвращает ошибку, если хранилище недоступно или неработоспособно.
	CheckStatus(ctx context.Context) error
}
