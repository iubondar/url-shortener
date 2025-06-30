// Package server предоставляет функциональность для запуска HTTP и gRPC серверов.
package server

import (
	"fmt"
	"net"
	"strings"

	gRPC "github.com/iubondar/url-shortener/internal/api/handlers/gRPC"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/iubondar/url-shortener/proto"
)

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

// Start запускает gRPC сервер.
// Возвращает ошибку, если сервер завершился с ошибкой.
func (s *GRPCServer) Start() error {
	// Создаем listener
	lis, err := net.Listen("tcp", s.config.GRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s.listener = lis

	zap.L().Debug("gRPC server started", zap.String("address", s.config.GRPCAddress))

	return s.server.Serve(lis)
}

// Shutdown выполняет graceful shutdown gRPC сервера
func (s *GRPCServer) Shutdown() error {
	// Graceful shutdown gRPC сервера
	s.server.GracefulStop()

	// Закрываем listener только если он еще не закрыт
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			// Игнорируем ошибку "use of closed network connection"
			if !strings.Contains(err.Error(), "use of closed network connection") {
				zap.L().Error("could not close listener", zap.Error(err))
				return err
			}
		}
	}

	return nil
}
