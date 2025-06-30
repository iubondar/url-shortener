package server

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gRPC "github.com/iubondar/url-shortener/internal/api/handlers/gRPC"
	"github.com/iubondar/url-shortener/internal/api/handlers/gRPC/mocks"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/proto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
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

func TestGRPCServerHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)
	service := gRPC.NewShortenerService(mockRepo, "http://localhost:8080", "")

	// Создаем gRPC сервер с interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.GRPCAuthInterceptor()),
	)
	proto.RegisterShortenerServer(grpcServer, service)

	// Создаем буферизованный listener для тестов
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	// Запускаем сервер
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// Создаем клиентское соединение
	ctx := context.Background()
	//nolint:staticcheck // grpc.DialContext is required for bufconn test setup
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	// Создаем клиент
	client := proto.NewShortenerClient(conn)

	// Тестируем Ping метод
	mockRepo.EXPECT().CheckStatus(gomock.Any()).Return(nil)

	resp, err := client.Ping(ctx, &proto.PingRequest{})
	assert.NoError(t, err)
	assert.Equal(t, proto.Status_STATUS_OK, resp.Status)
	assert.Empty(t, resp.Error)

	// Graceful shutdown
	grpcServer.GracefulStop()
	lis.Close()
}
