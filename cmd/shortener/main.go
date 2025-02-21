package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
)

func main() {
	config.Default.Load(os.Args[0], os.Args[1:])

	log.Fatal(
		http.ListenAndServe(
			config.Default.ServerAddress,
			router.Default(config.Default),
		),
	)
}
