package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/compress"
	"github.com/iubondar/url-shortener/internal/logging"
)

func NewRouter(config config.Config) (chi.Router, error) {

	repo, err := storage.NewFileRepository(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	createIDHandler := handlers.NewCreateIDHandler(repo, config.BaseURLAddress)
	shortenHandler := handlers.NewShortenHandler(repo, config.BaseURLAddress)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	r := chi.NewRouter()

	r.Use(logging.WithLogging, compress.WithGzipCompression)
	r.Post("/", createIDHandler.CreateID)
	r.Post("/api/shorten", shortenHandler.Shorten)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)

	return r, nil
}
