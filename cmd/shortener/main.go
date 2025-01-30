package main

import (
	"log"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
)

func main() {
	config.Default.Load()

	log.Fatal(
		http.ListenAndServe(
			config.Default.ServerAddress,
			router.Default(config.Default.BaseURLAddress),
		),
	)
}
