// Package router предоставляет функциональность для настройки маршрутизации HTTP-запросов.
// Использует библиотеку chi для создания маршрутизатора и настройки middleware.
package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/api/handlers"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/compress"
	"github.com/iubondar/url-shortener/internal/logging"
)

// NewRouter создает и настраивает маршрутизатор для обработки HTTP-запросов.
// Принимает базовый URL для формирования коротких ссылок и репозиторий для работы с хранилищем.
// Настраивает все необходимые маршруты и middleware:
//   - Логирование запросов
//   - Сжатие ответов
//   - Обработка создания коротких ссылок
//   - Обработка пакетного создания ссылок
//   - Получение списка ссылок пользователя
//   - Получение оригинального URL по короткому идентификатору
//   - Проверка доступности хранилища
//   - Удаление ссылок пользователя
//
// Возвращает настроенный маршрутизатор и ошибку, если она возникла.
func NewRouter(baseURL string, repo storage.Repository) (chi.Router, error) {
	createIDHandler := handlers.NewCreateIDHandler(repo, baseURL)
	shortenHandler := handlers.NewShortenHandler(repo, baseURL)
	shortenBatchHandler := handlers.NewShortenBatchHandler(repo, baseURL)
	userURLsHandler := handlers.NewUserUrlsHandler(repo, baseURL)
	retrieveURLHandler := handlers.NewRetrieveURLHandler(repo)
	pingHandler := handlers.NewPingHandler(repo)
	deleteURLsHandler := handlers.NewDeleteUrlsHandler(repo)

	r := chi.NewRouter()

	r.Use(logging.WithLogging, compress.WithGzipCompression)
	r.Post("/", createIDHandler.CreateID)
	r.Post("/api/shorten", shortenHandler.Shorten)
	r.Post("/api/shorten/batch", shortenBatchHandler.ShortenBatch)
	r.Get("/api/user/urls", userURLsHandler.RetrieveUserURLs)
	r.Get("/{id}", retrieveURLHandler.RetrieveURL)
	r.Get("/ping", pingHandler.Ping)
	r.Delete("/api/user/urls", deleteURLsHandler.DeleteUserURLs)

	// Подключаем pprof
	r.Mount("/debug/pprof", pprofRouter())

	return r, nil
}

// pprofRouter возвращает роутер с pprof-эндпоинтами
func pprofRouter() http.Handler {
	r := chi.NewRouter()

	// Регистрируем стандартные обработчики pprof
	r.Get("/", pprof.Index)
	r.Get("/cmdline", pprof.Cmdline)
	r.Get("/profile", pprof.Profile)
	r.Post("/symbol", pprof.Symbol)
	r.Get("/symbol", pprof.Symbol)
	r.Get("/trace", pprof.Trace)
	r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
	r.Get("/block", pprof.Handler("block").ServeHTTP)
	r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	return r
}
