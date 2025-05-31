package handlers

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/iubondar/url-shortener/internal/app/storage/file"
	"github.com/iubondar/url-shortener/internal/app/storage/pg"
	simple_storage "github.com/iubondar/url-shortener/internal/app/storage/simple"
)

type repository interface {
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
	RetrieveByShortURL(ctx context.Context, shortURL string) (record models.Record, err error)
	RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []models.Record, err error)
	DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string)
	CheckStatus(ctx context.Context) error
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)
}

// HandlerFactory определяет интерфейс для создания обработчиков HTTP-запросов.
// Фабрика инкапсулирует логику создания всех необходимых обработчиков,
// обеспечивая единую точку создания обработчиков в приложении.
type HandlerFactory interface {
	// CreateIDHandler создает обработчик для генерации короткого идентификатора URL
	CreateIDHandler() CreateIDHandler
	// ShortenHandler создает обработчик для сокращения URL
	ShortenHandler() ShortenHandler
	// ShortenBatchHandler создает обработчик для пакетного сокращения URL
	ShortenBatchHandler() ShortenBatchHandler
	// UserUrlsHandler создает обработчик для получения списка URL пользователя
	UserUrlsHandler() UserUrlsHandler
	// RetrieveURLHandler создает обработчик для получения оригинального URL по короткому идентификатору
	RetrieveURLHandler() RetrieveURLHandler
	// PingHandler создает обработчик для проверки доступности хранилища
	PingHandler() PingHandler
	// DeleteUrlsHandler создает обработчик для удаления URL пользователя
	DeleteUrlsHandler() DeleteUrlsHandler
}

// Factory реализует интерфейс HandlerFactory и создает обработчики HTTP-запросов.
// Фабрика использует репозиторий для работы с хранилищем данных и базовый URL
// для формирования коротких ссылок.
type Factory struct {
	repo    repository
	baseURL string
	db      *pg.DB
}

// NewFactory создает новую фабрику обработчиков на основе конфигурации приложения.
// Фабрика автоматически выбирает подходящий репозиторий в зависимости от конфигурации:
// - PostgreSQL, если указан DatabaseDSN
// - Файловое хранилище, если указан FileStoragePath
// - Простое хранилище в памяти в остальных случаях
func NewFactory(config config.Config) *Factory {
	var repo repository
	var db *pg.DB

	if len(config.DatabaseDSN) > 0 {
		var err error
		db, err = pg.NewDB(config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}

		repo, err = pg.NewPGRepository(db, 0)
		if err != nil {
			db.SQLDB.Close()
			log.Fatal(err)
		}
	} else if len(config.FileStoragePath) > 0 {
		var err error
		repo, err = file.NewFileRepository(config.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		repo = simple_storage.NewSimpleRepository()
	}
	return &Factory{repo: repo, baseURL: config.BaseURLAddress, db: db}
}

// Close освобождает ресурсы, используемые фабрикой.
// Должен быть вызван при завершении работы приложения.
func (f *Factory) Close() error {
	if f.db != nil {
		return f.db.SQLDB.Close()
	}
	return nil
}

// CreateIDHandler создает обработчик для генерации короткого идентификатора URL
func (f *Factory) CreateIDHandler() CreateIDHandler {
	return NewCreateIDHandler(f.repo, f.baseURL)
}

// ShortenHandler создает обработчик для сокращения URL
func (f *Factory) ShortenHandler() ShortenHandler {
	return NewShortenHandler(f.repo, f.baseURL)
}

// ShortenBatchHandler создает обработчик для пакетного сокращения URL
func (f *Factory) ShortenBatchHandler() ShortenBatchHandler {
	return NewShortenBatchHandler(f.repo, f.baseURL)
}

// UserUrlsHandler создает обработчик для получения списка URL пользователя
func (f *Factory) UserUrlsHandler() UserUrlsHandler {
	return NewUserUrlsHandler(f.repo, f.baseURL)
}

// RetrieveURLHandler создает обработчик для получения оригинального URL по короткому идентификатору
func (f *Factory) RetrieveURLHandler() RetrieveURLHandler {
	return NewRetrieveURLHandler(f.repo)
}

// PingHandler создает обработчик для проверки доступности хранилища
func (f *Factory) PingHandler() PingHandler {
	return NewPingHandler(f.repo)
}

// DeleteUrlsHandler создает обработчик для удаления URL пользователя
func (f *Factory) DeleteUrlsHandler() DeleteUrlsHandler {
	return NewDeleteUrlsHandler(f.repo)
}
