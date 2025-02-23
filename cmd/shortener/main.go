package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
)

func main() {
	config, err := config.NewConfig(os.Args[0], os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	router, err := router.NewRouter(*config)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe(config.ServerAddress, router),
	)
}
