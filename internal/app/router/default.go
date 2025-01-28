package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

func Default(baseUrl string) chi.Router {

	repo := storage.NewSimpleRepository()

	createIdHandler := handlers.NewCreateIdHandler(repo, baseUrl)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	r := chi.NewRouter()

	r.Post("/", createIdHandler.CreateId)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)

	return r
}
