package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
)

// ShortenIn представляет входные данные для создания сокращенного URL.
type ShortenIn struct {
	URL string `json:"url"` // оригинальный URL для сокращения
}

// ShortenOut представляет выходные данные создания сокращенного URL.
type ShortenOut struct {
	Result string `json:"result"` // сокращенный URL
}

// ShortenHandler обрабатывает запросы на создание сокращенного URL.
// Позволяет создать сокращенную ссылку для одного URL.
type ShortenHandler struct {
	saver   URLSaver // репозиторий для хранения URL
	baseURL string   // базовый URL для формирования сокращенных ссылок
}

// NewShortenHandler создает новый экземпляр ShortenHandler.
// Принимает репозиторий для хранения URL и базовый URL для формирования сокращенных ссылок.
func NewShortenHandler(saver URLSaver, baseURL string) ShortenHandler {
	return ShortenHandler{
		saver:   saver,
		baseURL: baseURL,
	}
}

// Shorten обрабатывает HTTP POST запрос для создания сокращенного URL.
// Принимает URL в теле запроса в формате JSON.
// Возвращает сокращенный URL в формате JSON.
// Возвращает статус 201 Created для нового URL или 409 Conflict если URL уже существует.
func (handler ShortenHandler) Shorten(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var in ShortenIn
	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &in); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := url.ParseRequestURI(in.URL)
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

	baseURL := strings.TrimSuffix(strings.TrimPrefix(handler.baseURL, "http://"), "/")
	out := ShortenOut{
		Result: fmt.Sprintf("http://%s/%s", baseURL, id),
	}

	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if exists {
		res.WriteHeader(http.StatusConflict)
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	res.Write(resp)
}
