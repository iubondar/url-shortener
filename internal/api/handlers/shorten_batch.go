package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/storage"
)

// ShortenBatchIn представляет входные данные для пакетного создания сокращенных URL.
type ShortenBatchIn struct {
	CorrelationID string `json:"correlation_id"` // идентификатор для связи с оригинальным URL
	OriginalURL   string `json:"original_url"`   // оригинальный URL для сокращения
}

// ShortenBatchOut представляет выходные данные пакетного создания сокращенных URL.
type ShortenBatchOut struct {
	CorrelationID string `json:"correlation_id"` // идентификатор для связи с оригинальным URL
	ShortURL      string `json:"short_url"`      // сокращенный URL
}

// ShortenBatchHandler обрабатывает запросы на пакетное создание сокращенных URL.
// Позволяет создать несколько сокращенных URL за один запрос.
type ShortenBatchHandler struct {
	repo    storage.Repository // репозиторий для хранения URL
	baseURL string             // базовый URL для формирования сокращенных ссылок
}

// NewShortenBatchHandler создает новый экземпляр ShortenBatchHandler.
// Принимает репозиторий для хранения URL и базовый URL для формирования сокращенных ссылок.
func NewShortenBatchHandler(repo storage.Repository, baseURL string) ShortenBatchHandler {
	return ShortenBatchHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

// ShortenBatch обрабатывает HTTP POST запрос для пакетного создания сокращенных URL.
// Принимает массив URL в теле запроса в формате JSON.
// Возвращает массив созданных сокращенных URL в формате JSON.
// Возвращает статус 201 Created в случае успеха.
func (handler ShortenBatchHandler) ShortenBatch(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var in []ShortenBatchIn
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
	urls := make([]string, 0, len(in))
	for _, elem := range in {
		// Проверяем URL
		URL, err := url.ParseRequestURI(elem.OriginalURL)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		urls = append(urls, URL.String())
	}

	ids, err := handler.repo.SaveURLs(req.Context(), urls)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := make([]ShortenBatchOut, 0, len(in))
	baseURL := strings.TrimSuffix(strings.TrimPrefix(handler.baseURL, "http://"), "/")
	for i := 0; i < len(in); i++ {
		outElem := ShortenBatchOut{
			CorrelationID: in[i].CorrelationID,
			ShortURL:      fmt.Sprintf("http://%s/%s", baseURL, ids[i])}
		out = append(out, outElem)
	}

	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)
}
