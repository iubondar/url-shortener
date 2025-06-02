package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

// URLBatchSaver определяет интерфейс для пакетного сохранения URL в хранилище.
type URLBatchSaver interface {
	// SaveURLs сохраняет массив URL в хранилище.
	// Возвращает массив коротких идентификаторов и ошибку.
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)
}

// ShortenBatchHandler обрабатывает запросы на пакетное создание сокращенных URL.
// Позволяет создать несколько сокращенных URL за один запрос.
type ShortenBatchHandler struct {
	saver   URLBatchSaver // репозиторий для хранения URL
	baseURL string        // базовый URL для формирования сокращенных ссылок
}

// NewShortenBatchHandler создает новый экземпляр ShortenBatchHandler.
// Принимает репозиторий для хранения URL и базовый URL для формирования сокращенных ссылок.
func NewShortenBatchHandler(saver URLBatchSaver, baseURL string) ShortenBatchHandler {
	return ShortenBatchHandler{
		saver:   saver,
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
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
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

	ids, err := handler.saver.SaveURLs(req.Context(), urls)
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

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(out); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
