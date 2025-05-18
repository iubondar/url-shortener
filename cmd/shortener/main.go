// Package main предоставляет серверную часть сервиса сокращения URL.
// Сервер принимает длинные URL, генерирует для них короткие идентификаторы
// и сохраняет соответствия в выбранном хранилище (память, файл или база данных).
package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
	"github.com/iubondar/url-shortener/internal/app/storage"

	_ "net/http/pprof" // подключаем пакет pprof
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

// main является точкой входа в серверное приложение.
// Функция инициализирует конфигурацию, подключает выбранное хранилище данных,
// настраивает маршрутизацию и запускает HTTP-сервер.
func main() {
	config, err := config.NewConfig(os.Args[0], os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	zap.L().Sugar().Debugln(
		"Config: ",
		"ServerAddress", config.ServerAddress,
		"BaseURLAddress", config.BaseURLAddress,
		"FileStoragePath", config.FileStoragePath,
		"DatabaseDSN", config.DatabaseDSN,
	)

	var repo storage.Repository

	if len(config.DatabaseDSN) > 0 {
		db, err := storage.NewDB(config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}

		defer db.SQLDB.Close()

		repo = db.Repo
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

	zap.L().Sugar().Debugln("Starting serving requests: ", config.ServerAddress)
	log.Fatal(
		http.ListenAndServe(config.ServerAddress, router),
	)
}
