package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

type ShortenIn struct {
	URL string `json:"url"`
}

type ShortenOut struct {
	Result string `json:"result"`
}

type ShortenHandler struct {
	repo    storage.Repository
	baseURL string
}

func NewShortenHandler(repo storage.Repository, baseURL string) ShortenHandler {
	return ShortenHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

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

	userID, err := auth.SetAuthCookie(res, req)
	if err != nil {
		http.Error(res, "Error setting userID "+err.Error(), http.StatusBadRequest)
		return
	}

	id, exists, err := handler.repo.SaveURL(req.Context(), userID, url.String())
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
