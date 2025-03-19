package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

type UserUrlsHandler struct {
	repo storage.Repository
}

func NewUserUrlsHandler(repo storage.Repository) UserUrlsHandler {
	return UserUrlsHandler{
		repo: repo,
	}
}

func (handler UserUrlsHandler) RetrieveUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	authCookie, err := req.Cookie(auth.AuthCookieName)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(authCookie.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	URLPairs, err := handler.repo.RetrieveUserURLs(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(URLPairs)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(URLPairs) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
	}

	res.Write(resp)
}
