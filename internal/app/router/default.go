package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/compress"
	"github.com/iubondar/url-shortener/internal/logging"
)

func Default(baseURL string) chi.Router {

	repo := storage.NewSimpleRepository()

	createIDHandler := handlers.NewCreateIDHandler(repo, baseURL)
	shortenHandler := handlers.NewShortenHandler(repo, baseURL)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	r := chi.NewRouter()

	r.Use(logging.WithLogging, compress.WithGzipCompression)
	r.Post("/", createIDHandler.CreateID)
	r.Post("/api/shorten", shortenHandler.Shorten)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)

	return r
}
