package main

import (
	"log"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/router"
)

func main() {
	log.Fatal(http.ListenAndServe(`localhost:8080`, router.DefaultRouter()))
}
