package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

func main() {
	config, err := config.NewConfig(os.Args[0], os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fileRepo, err := storage.NewFileRepository(config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	repo, err := storage.NewPGRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	router, err := router.NewRouter(config.BaseURLAddress, fileRepo, repo)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe(config.ServerAddress, router),
	)
}
