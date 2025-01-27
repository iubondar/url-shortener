package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/storage"
)

func DefaultRouter() chi.Router {

	repo := storage.NewSimpleRepository()

	createIdHandler := handlers.NewCreateIdHandler(repo)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	r := chi.NewRouter()

	r.Post("/", createIdHandler.CreateId)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)

	return r
}
