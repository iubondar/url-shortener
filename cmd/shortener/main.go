// Package main предоставляет серверную часть сервиса сокращения URL.
// Сервер принимает длинные URL, генерирует для них короткие идентификаторы
// и сохраняет соответствия в выбранном хранилище (память, файл или база данных).
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/router"
	"github.com/iubondar/url-shortener/internal/app/server"

	_ "net/http/pprof" // подключаем пакет pprof
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

// main является точкой входа в серверное приложение.
// Функция инициализирует конфигурацию, подключает выбранное хранилище данных,
// настраивает маршрутизацию и запускает HTTP-сервер.
func main() {
	printVersion()

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
		"EnableHTTPS", config.EnableHTTPS,
	)

	factory := handlers.NewFactory(config)
	defer func() {
		if err := factory.Close(); err != nil {
			zap.L().Sugar().Errorf("Error closing factory: %v", err)
		}
	}()

	router, err := router.NewRouter(factory)
	if err != nil {
		log.Fatal(err)
	}

	srv := server.New(config, router)
	grpcServer, err := server.NewGRPCServer(config, factory.ShortenerService())
	if err != nil {
		log.Fatal(err)
	}

	// Каналы для обработки ошибок серверов
	httpErrors := make(chan error, 1)
	grpcErrors := make(chan error, 1)

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		if err := srv.Start(); err != nil {
			httpErrors <- err
		}
	}()

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		if err := grpcServer.Start(); err != nil {
			grpcErrors <- err
		}
	}()

	// Канал для обработки сигналов завершения от ОС
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ожидаем либо ошибку одного из серверов, либо сигнал завершения
	select {
	case err := <-httpErrors:
		zap.L().Sugar().Errorf("HTTP server error: %v", err)
		// При ошибке HTTP сервера завершаем gRPC сервер
		if err := grpcServer.Shutdown(); err != nil {
			zap.L().Sugar().Errorf("Error shutting down gRPC server: %v", err)
		}
		log.Fatal(err)

	case err := <-grpcErrors:
		zap.L().Sugar().Errorf("gRPC server error: %v", err)
		// При ошибке gRPC сервера завершаем HTTP сервер
		if err := srv.Shutdown(); err != nil {
			zap.L().Sugar().Errorf("Error shutting down HTTP server: %v", err)
		}
		log.Fatal(err)

	case sig := <-shutdown:
		zap.L().Sugar().Infof("Received shutdown signal: %v", sig)
		// Graceful shutdown обоих серверов
		if err := srv.Shutdown(); err != nil {
			zap.L().Sugar().Errorf("Error shutting down HTTP server: %v", err)
		}
		if err := grpcServer.Shutdown(); err != nil {
			zap.L().Sugar().Errorf("Error shutting down gRPC server: %v", err)
		}
	}
}

func printVersion() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
