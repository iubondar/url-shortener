package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/compress"
	"github.com/iubondar/url-shortener/internal/logging"
)

func NewRouter(BaseURL string, repo storage.Repository) (chi.Router, error) {
	createIDHandler := handlers.NewCreateIDHandler(repo, BaseURL)
	shortenHandler := handlers.NewShortenHandler(repo, BaseURL)
	shortenBatchHandler := handlers.NewShortenBatchHandler(repo, BaseURL)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)
	pingHandler := handlers.NewPingHandler(repo)

	r := chi.NewRouter()

	r.Use(logging.WithLogging, compress.WithGzipCompression)
	r.Post("/", createIDHandler.CreateID)
	r.Post("/api/shorten", shortenHandler.Shorten)
	r.Post("/api/shorten/batch", shortenBatchHandler.ShortenBatch)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)
	r.Get("/ping", pingHandler.Ping)

	return r, nil
}
