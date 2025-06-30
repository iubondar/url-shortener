// Package server предоставляет функциональность для запуска HTTP и gRPC серверов.
package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	gRPC "github.com/iubondar/url-shortener/internal/api/handlers/gRPC"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/proto"
)

const gRPCPort = ":3200"

// GRPCServer представляет gRPC сервер приложения.
type GRPCServer struct {
	config   config.Config
	server   *grpc.Server
	listener net.Listener
}

// NewGRPCServer создает новый экземпляр GRPCServer.
// Принимает конфигурацию и фабрику обработчиков.
func NewGRPCServer(config config.Config, service *gRPC.ShortenerService) (*GRPCServer, error) {
	// Создаем gRPC сервер с аутентификационным interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.GRPCAuthInterceptor()),
	)

	// Регистрируем gRPC сервисы
	proto.RegisterShortenerServer(grpcServer, service)

	return &GRPCServer{
		config: config,
		server: grpcServer,
	}, nil
}

// Start запускает gRPC сервер в отдельной горутине.
// Возвращает ошибку, если сервер завершился с ошибкой.
func (s *GRPCServer) Start() error {
	// Создаем listener
	lis, err := net.Listen("tcp", s.config.ServerAddress+gRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s.listener = lis

	// Канал для обработки ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := s.server.Serve(lis); err != nil {
			serverErrors <- err
		}
	}()

	// Канал для обработки сигналов завершения от ОС
	shutdown := make(chan os.Signal, 1)
	// Регистрируем обработчики сигналов
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ожидаем либо ошибку сервера, либо сигнал завершения
	select {
	case err := <-serverErrors:
		zap.L().Error("gRPC server error", zap.Error(err))
		return err

	case sig := <-shutdown:
		zap.L().Info("start gRPC shutdown", zap.String("signal", sig.String()))
		return s.Shutdown()
	}
}

// Shutdown выполняет graceful shutdown gRPC сервера
func (s *GRPCServer) Shutdown() error {
	// Graceful shutdown gRPC сервера
	s.server.GracefulStop()

	// Закрываем listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			zap.L().Error("could not close listener", zap.Error(err))
			return err
		}
	}

	return nil
}
