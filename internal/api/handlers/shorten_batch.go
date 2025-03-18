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

type ShortenBatchIn struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchOut struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ShortenBatchHandler struct {
	repo    storage.Repository
	baseURL string
}

func NewShortenBatchHandler(repo storage.Repository, baseURL string) ShortenBatchHandler {
	return ShortenBatchHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

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
