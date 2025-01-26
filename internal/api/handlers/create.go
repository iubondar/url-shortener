package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/iubondar/url-shortener/internal/storage"
)

type CreateIdHandler struct {
	repo storage.Repository
}

func NewCreateIdHandler(repo storage.Repository) CreateIdHandler {
	return CreateIdHandler{
		repo: repo,
	}
}

func (handler CreateIdHandler) CreateId(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	url := string(body)

	id, exists, err := handler.repo.SaveURL(url)
	if err != nil {
		http.Error(res, "Can't save URL", http.StatusBadRequest)
		return
	}

	if !exists {
		res.WriteHeader(http.StatusCreated)
	}

	result := fmt.Sprintf("http://%s/%s", req.Host, id)

	res.Write([]byte(result))
}
