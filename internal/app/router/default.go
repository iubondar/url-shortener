package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/logging"
)

func Default(baseURL string) chi.Router {

	repo := storage.NewSimpleRepository()

	CreateIDHandler := handlers.NewCreateIDHandler(repo, baseURL)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	r := chi.NewRouter()

	r.Use(logging.WithLogging)
	r.Post("/", CreateIDHandler.CreateID)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)

	return r
}
