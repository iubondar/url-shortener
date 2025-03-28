package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

type CreateIDHandler struct {
	repo    storage.Repository
	baseURL string
}

func NewCreateIDHandler(repo storage.Repository, baseURL string) CreateIDHandler {
	return CreateIDHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

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

	id, exists, err := handler.repo.SaveURL(req.Context(), userID, url.String())
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

	res.Write([]byte(result))
}
