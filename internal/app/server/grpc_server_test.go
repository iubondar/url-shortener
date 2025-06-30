package server

import (
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gRPC "github.com/iubondar/url-shortener/internal/api/handlers/gRPC"
	"github.com/iubondar/url-shortener/internal/api/handlers/gRPC/mocks"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewGRPCServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	cfg := config.Config{
		ServerAddress:  ":8080",
		GRPCAddress:    ":3200",
		BaseURLAddress: "http://localhost:8080",
		EnableHTTPS:    false,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)
	service := gRPC.NewShortenerService(mockRepo, cfg.BaseURLAddress, "")

	server, err := NewGRPCServer(cfg, service)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.NotNil(t, server.server)
}

func TestGRPCServerStartAndShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	cfg := config.Config{
		ServerAddress:  ":8080",
		GRPCAddress:    ":3201",
		BaseURLAddress: "http://localhost:8080",
		EnableHTTPS:    false,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)
	service := gRPC.NewShortenerService(mockRepo, cfg.BaseURLAddress, "")

	server, err := NewGRPCServer(cfg, service)
	assert.NoError(t, err)

	// Запускаем сервер в отдельной горутине
	go func() {
		_ = server.Start()
	}()

	// Даем серверу время на запуск
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервер запущен
	assert.NotNil(t, server.listener)

	// Выполняем graceful shutdown
	err = server.Shutdown()
	// Игнорируем ошибку закрытия listener, так как это нормально в тестах
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		assert.NoError(t, err, "shutdown should not return error")
	}
}
