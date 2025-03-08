package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	config, err := config.NewConfig(os.Args[0], os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var repo storage.Repository

	if len(config.DatabaseDSN) > 0 {
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		repo, err = storage.NewPGRepository(context.Background(), db)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(config.FileStoragePath) > 0 {
		repo, err = storage.NewFileRepository(config.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		repo = storage.NewSimpleRepository()
	}

	router, err := router.NewRouter(config.BaseURLAddress, repo)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe(config.ServerAddress, router),
	)
}
