package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

type UserUrlsHandler struct {
	repo    storage.Repository
	baseURL string
}

func NewUserUrlsHandler(repo storage.Repository, baseURL string) UserUrlsHandler {
	return UserUrlsHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

type UserUrlsOut struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (handler UserUrlsHandler) RetrieveUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.SetAuthCookie(res, req)
	if err != nil {
		http.Error(res, "Error setting userID "+err.Error(), http.StatusBadRequest)
		return
	}

	URLPairs, err := handler.repo.RetrieveUserURLs(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := make([]UserUrlsOut, 0, len(URLPairs))
	baseURL := strings.TrimSuffix(strings.TrimPrefix(handler.baseURL, "http://"), "/")
	for i := 0; i < len(URLPairs); i++ {
		outElem := UserUrlsOut{
			ShortURL:    fmt.Sprintf("http://%s/%s", baseURL, URLPairs[i].ID),
			OriginalURL: URLPairs[i].URL,
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
