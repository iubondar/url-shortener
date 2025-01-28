package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
)

func main() {
	flag.Parse()

	log.Fatal(
		http.ListenAndServe(
			config.Default.ServerAddress,
			router.Default(config.Default.BaseURLAddress),
		),
	)
}
