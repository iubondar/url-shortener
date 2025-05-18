package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

// UserUrlsHandler обрабатывает запросы на получение списка сокращенных URL пользователя.
// Позволяет пользователю получить список всех своих сокращенных URL.
type UserUrlsHandler struct {
	repo    storage.Repository // репозиторий для хранения URL
	baseURL string             // базовый URL для формирования сокращенных ссылок
}

// NewUserUrlsHandler создает новый экземпляр UserUrlsHandler.
// Принимает репозиторий для хранения URL и базовый URL для формирования сокращенных ссылок.
func NewUserUrlsHandler(repo storage.Repository, baseURL string) UserUrlsHandler {
	return UserUrlsHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

// UserUrlsOut представляет выходные данные для списка URL пользователя.
type UserUrlsOut struct {
	ShortURL    string `json:"short_url"`    // сокращенный URL
	OriginalURL string `json:"original_url"` // оригинальный URL
}

// RetrieveUserURLs обрабатывает HTTP GET запрос для получения списка сокращенных URL пользователя.
// Возвращает список сокращенных URL в формате JSON.
// Возвращает статус 200 OK если есть URL, 204 No Content если список пуст.
func (handler UserUrlsHandler) RetrieveUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromAuthCookieOrSetNew(res, req)
	if err != nil {
		http.Error(res, "Error setting userID "+err.Error(), http.StatusBadRequest)
		return
	}

	records, err := handler.repo.RetrieveUserURLs(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := make([]UserUrlsOut, 0, len(records))
	baseURL := strings.TrimSuffix(strings.TrimPrefix(handler.baseURL, "http://"), "/")
	for i := 0; i < len(records); i++ {
		outElem := UserUrlsOut{
			ShortURL:    fmt.Sprintf("http://%s/%s", baseURL, records[i].ShortURL),
			OriginalURL: records[i].OriginalURL,
		}
		out = append(out, outElem)
	}
	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(out) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
	}

	res.Write(resp)
}
