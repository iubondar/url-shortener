package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

// DeleteUrlsHandler обрабатывает запросы на удаление сокращенных URL.
// Позволяет пользователю удалить свои сокращенные ссылки.
type DeleteUrlsHandler struct {
	repo storage.Repository // репозиторий для хранения URL
}

// NewDeleteUrlsHandler создает новый экземпляр DeleteUrlsHandler.
// Принимает репозиторий для хранения URL.
func NewDeleteUrlsHandler(repo storage.Repository) DeleteUrlsHandler {
	return DeleteUrlsHandler{
		repo: repo,
	}
}

// DeleteUserURLs обрабатывает HTTP DELETE запрос для удаления сокращенных URL.
// Принимает массив сокращенных URL в теле запроса в формате JSON.
// Удаляет только те URL, которые принадлежат текущему пользователю.
// Возвращает статус 202 Accepted в случае успеха.
func (handler DeleteUrlsHandler) DeleteUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(res, "Only DELETE requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromAuthCookieOrSetNew(res, req)
	if err != nil {
		http.Error(res, "Error setting userID "+err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	// читаем тело запроса
	_, err = buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// десериализуем JSON
	var shortURLs []string
	if err = json.Unmarshal(buf.Bytes(), &shortURLs); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// запрос на удаление
	handler.repo.DeleteByShortURLs(req.Context(), userID, shortURLs)

	// сразу возвращаем статус
	res.WriteHeader(http.StatusAccepted)
}
