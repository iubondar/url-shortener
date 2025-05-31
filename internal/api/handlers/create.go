package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
)

// CreateIDHandler обрабатывает запросы на создание сокращенных URL.
type CreateIDHandler struct {
	saver   URLSaver // репозиторий для хранения URL
	baseURL string   // базовый URL для формирования сокращенных ссылок
}

// NewCreateIDHandler создает новый экземпляр CreateIDHandler.
// Принимает репозиторий для хранения URL и базовый URL для формирования сокращенных ссылок.
func NewCreateIDHandler(saver URLSaver, baseURL string) CreateIDHandler {
	return CreateIDHandler{
		saver:   saver,
		baseURL: baseURL,
	}
}

// CreateID обрабатывает HTTP POST запрос для создания сокращенного URL.
// Принимает URL в теле запроса, проверяет его валидность и сохраняет в репозитории.
// Возвращает сокращенный URL в формате "http://{baseURL}/{id}".
// В случае успеха возвращает статус 201 Created, если URL уже существует - 409 Conflict.
func (handler CreateIDHandler) CreateID(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	url, err := url.ParseRequestURI(string(body))
	if err != nil {
		http.Error(res, "URL is not valid", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromAuthCookieOrSetNew(res, req)
	if err != nil {
		http.Error(res, "Error setting userID "+err.Error(), http.StatusBadRequest)
		return
	}

	id, exists, err := handler.saver.SaveURL(req.Context(), userID, url.String())
	if err != nil {
		http.Error(res, "Can't save URL", http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "text/plain")

	if exists {
		res.WriteHeader(http.StatusConflict)
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	baseURL := strings.TrimSuffix(strings.TrimPrefix(handler.baseURL, "http://"), "/")
	result := fmt.Sprintf("http://%s/%s", baseURL, id)

	if _, err := res.Write([]byte(result)); err != nil {
		http.Error(res, "Error writing response", http.StatusInternalServerError)
		return
	}
}
