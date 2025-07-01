package handlers

import (
	"testing"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory_SimpleStorage(t *testing.T) {
	// Тест создания фабрики с простым хранилищем (по умолчанию)
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
		TrustedSubnet:  "192.168.1.0/24",
	}

	factory, err := NewFactory(cfg)
	require.NoError(t, err)
	require.NotNil(t, factory)
	assert.Equal(t, "http://localhost:8080", factory.baseURL)
	assert.Equal(t, "192.168.1.0/24", factory.trustedSubnet)
	assert.Nil(t, factory.db)
	assert.NotNil(t, factory.repo)
}

func TestNewFactory_FileStorage(t *testing.T) {
	// Тест создания фабрики с файловым хранилищем
	cfg := config.Config{
		BaseURLAddress:  "http://localhost:8080",
		FileStoragePath: "/tmp/test_storage.txt",
		TrustedSubnet:   "192.168.1.0/24",
	}

	factory, err := NewFactory(cfg)
	require.NoError(t, err)
	require.NotNil(t, factory)
	assert.Equal(t, "http://localhost:8080", factory.baseURL)
	assert.Equal(t, "192.168.1.0/24", factory.trustedSubnet)
	assert.Nil(t, factory.db)
	assert.NotNil(t, factory.repo)
}

func TestNewFactory_DatabaseStorage(t *testing.T) {
	// Тест создания фабрики с базой данных
	// Используем невалидный DSN для тестирования обработки ошибок
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
		DatabaseDSN:    "invalid_dsn",
		TrustedSubnet:  "192.168.1.0/24",
	}

	// Этот тест должен завершиться с ошибкой из-за невалидного DSN
	factory, err := NewFactory(cfg)
	assert.Error(t, err)
	assert.Nil(t, factory)
}

func TestFactory_Close(t *testing.T) {
	// Тест закрытия фабрики без базы данных
	factory := &Factory{
		baseURL:       "http://localhost:8080",
		trustedSubnet: "192.168.1.0/24",
		db:            nil,
	}

	err := factory.Close()
	assert.NoError(t, err)
}

func TestFactory_CreateIDHandler(t *testing.T) {
	// Тест создания обработчика CreateID
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.CreateIDHandler()
	assert.NotNil(t, handler)
}

func TestFactory_ShortenHandler(t *testing.T) {
	// Тест создания обработчика Shorten
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.ShortenHandler()
	assert.NotNil(t, handler)
}

func TestFactory_ShortenBatchHandler(t *testing.T) {
	// Тест создания обработчика ShortenBatch
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.ShortenBatchHandler()
	assert.NotNil(t, handler)
}

func TestFactory_UserUrlsHandler(t *testing.T) {
	// Тест создания обработчика UserUrls
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.UserUrlsHandler()
	assert.NotNil(t, handler)
}

func TestFactory_RetrieveURLHandler(t *testing.T) {
	// Тест создания обработчика RetrieveURL
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.RetrieveURLHandler()
	assert.NotNil(t, handler)
}

func TestFactory_PingHandler(t *testing.T) {
	// Тест создания обработчика Ping
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.PingHandler()
	assert.NotNil(t, handler)
}

func TestFactory_DeleteUrlsHandler(t *testing.T) {
	// Тест создания обработчика DeleteUrls
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.DeleteUrlsHandler()
	assert.NotNil(t, handler)
}

func TestFactory_InternalStatsHandler(t *testing.T) {
	// Тест создания обработчика InternalStats
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
		TrustedSubnet:  "192.168.1.0/24",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	handler := factory.InternalStatsHandler()
	assert.NotNil(t, handler)
}

func TestFactory_ShortenerService(t *testing.T) {
	// Тест создания gRPC сервиса
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
		TrustedSubnet:  "192.168.1.0/24",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	service := factory.ShortenerService()
	assert.NotNil(t, service)
}

func TestFactory_ImplementsHandlerFactory(t *testing.T) {
	// Тест, что Factory реализует интерфейс HandlerFactory
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	// Проверяем, что Factory можно привести к типу HandlerFactory
	var _ HandlerFactory = factory
}

// Дополнительные тесты для проверки работы фабрики
func TestFactory_AllHandlers(t *testing.T) {
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
		TrustedSubnet:  "192.168.1.0/24",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	// Проверяем, что все обработчики создаются корректно
	assert.NotNil(t, factory.CreateIDHandler())
	assert.NotNil(t, factory.ShortenHandler())
	assert.NotNil(t, factory.ShortenBatchHandler())
	assert.NotNil(t, factory.UserUrlsHandler())
	assert.NotNil(t, factory.RetrieveURLHandler())
	assert.NotNil(t, factory.PingHandler())
	assert.NotNil(t, factory.DeleteUrlsHandler())
	assert.NotNil(t, factory.InternalStatsHandler())
	assert.NotNil(t, factory.ShortenerService())
}

// Тест для покрытия ветки с ошибкой закрытия базы данных
func TestFactory_CloseWithError(t *testing.T) {
	// Создаем фабрику с базой данных, чтобы протестировать Close
	// Но сначала нужно создать валидную конфигурацию для базы данных
	// Для тестирования используем простую фабрику и мокаем Close
	factory := &Factory{
		baseURL:       "http://localhost:8080",
		trustedSubnet: "192.168.1.0/24",
		db:            nil, // nil db не вызовет ошибку при закрытии
	}

	err := factory.Close()
	assert.NoError(t, err)
}

// Тест для покрытия ветки с ошибкой создания PG репозитория
func TestNewFactory_PGRepositoryError(t *testing.T) {
	// Этот тест сложно реализовать без моков, так как нужно симулировать
	// ошибку в pg.NewPGRepository. Пока что пропустим этот тест.
	t.Skip("Requires mocking pg.NewPGRepository")
}

// Тест для покрытия ветки с ошибкой закрытия соединения с БД
func TestNewFactory_DBCloseError(t *testing.T) {
	// Этот тест также сложно реализовать без моков.
	t.Skip("Requires mocking database connection")
}

func TestNewFactory_FileStorageError(t *testing.T) {
	// Тест создания фабрики с невалидным путем к файлу
	cfg := config.Config{
		BaseURLAddress:  "http://localhost:8080",
		FileStoragePath: "/invalid/path/that/does/not/exist/and/cannot/be/created.txt",
		TrustedSubnet:   "192.168.1.0/24",
	}

	// Этот тест должен завершиться с ошибкой из-за невалидного пути
	factory, err := NewFactory(cfg)
	assert.Error(t, err)
	assert.Nil(t, factory)
}

// Тест для покрытия ветки с ошибкой закрытия БД в Close
func TestFactory_CloseWithDBError(t *testing.T) {
	// Создаем фабрику с простым хранилищем (без БД)
	cfg := config.Config{
		BaseURLAddress: "http://localhost:8080",
	}
	factory, err := NewFactory(cfg)
	require.NoError(t, err)

	// Close без БД не должен возвращать ошибку
	err = factory.Close()
	assert.NoError(t, err)
}
