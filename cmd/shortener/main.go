package main

import (
	"log"
	"net/http"

	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/storage"
)

func main() {
	repo := storage.NewSimpleRepository()

	createIdHandler := handlers.NewCreateIdHandler(repo)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, createIdHandler.CreateId)
	mux.HandleFunc(`/{id}`, retrieveURLHandler.RetrieveURL)

	log.Fatal(http.ListenAndServe(`localhost:8080`, mux))
}
