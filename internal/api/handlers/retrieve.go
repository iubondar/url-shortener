package handlers

import (
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/storage"
)

type RetrieveURLHandler struct {
	repo storage.Repository
}

func NewRetrieveURLHandler(repo storage.Repository) RetrieveURLHandler {
	return RetrieveURLHandler{
		repo: repo,
	}
}

func (handler RetrieveURLHandler) RetrieveURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := req.PathValue("id")
	if len(id) == 0 {
		http.Error(res, "Can't find id parameter in query path", http.StatusBadRequest)
		return
	}

	url, err := handler.repo.RetrieveURL(id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Add("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
