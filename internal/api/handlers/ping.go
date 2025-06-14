package handlers

import (
	"context"
	"net/http"
)

// StatusChecker определяет интерфейс для проверки состояния хранилища.
// Используется для проверки доступности и работоспособности хранилища.
type StatusChecker interface {
	// CheckStatus проверяет состояние хранилища.
	// Возвращает ошибку, если хранилище недоступно или неработоспособно.
	CheckStatus(ctx context.Context) error
}

// PingHandler обрабатывает запросы для проверки доступности сервиса.
// Используется для проверки работоспособности сервера и его подключения к хранилищу.
type PingHandler struct {
	checker StatusChecker // интерфейс для проверки статуса хранилища
}

// NewPingHandler создает новый экземпляр PingHandler.
// Принимает интерфейс для проверки статуса хранилища.
func NewPingHandler(checker StatusChecker) PingHandler {
	return PingHandler{
		checker: checker,
	}
}

// Ping обрабатывает HTTP GET запрос для проверки доступности сервиса.
// Проверяет подключение к хранилищу данных.
// Возвращает статус 200 OK в случае успеха или 500 Internal Server Error при ошибке.
func (handler PingHandler) Ping(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	err := handler.checker.CheckStatus(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
