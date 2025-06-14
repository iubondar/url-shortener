// Package server предоставляет функциональность для запуска HTTP и HTTPS серверов.
package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"github.com/iubondar/url-shortener/internal/app/config"
)

// Server представляет HTTP/HTTPS сервер приложения.
type Server struct {
	config config.Config
	router http.Handler
	server *http.Server
}

// New создает новый экземпляр Server.
// Принимает конфигурацию и HTTP-роутер.
func New(config config.Config, router http.Handler) *Server {
	return &Server{
		config: config,
		router: router,
	}
}

// Start запускает HTTP или HTTPS сервер в отдельной горутине.
// Если EnableHTTPS=true, запускается HTTPS сервер с автоматическим получением сертификатов
// или использованием локальных сертификатов для localhost/IP.
// Возвращает ошибку, если сервер завершился с ошибкой.
func (s *Server) Start() error {
	zap.L().Sugar().Debugln("Starting serving requests: ", s.config.ServerAddress)

	// Канал для обработки ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if s.config.EnableHTTPS {
			serverErrors <- s.startHTTPServerTLS()
		}
		serverErrors <- s.startHTTPServer()
	}()

	// Канал для обработки сигналов завершения от ОС
	shutdown := make(chan os.Signal, 1)
	// Регистрируем обработчики сигналов
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ожидаем либо ошибку сервера, либо сигнал завершения
	select {
	case err := <-serverErrors:
		zap.L().Error("server error", zap.Error(err))
		return err

	case sig := <-shutdown:
		zap.L().Info("start shutdown", zap.String("signal", sig.String()))
		return s.Shutdown()
	}

}

// Shutdown выполняет graceful shutdown сервера
func (s *Server) Shutdown() error {
	// Устанавливаем таймаут 5 секунд для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся корректно завершить работу сервера
	if err := s.server.Shutdown(ctx); err != nil {
		zap.L().Error("graceful shutdown did not complete", zap.Error(err))
		// Если плавное завершение не удалось, принудительно закрываем сервер
		if err := s.server.Close(); err != nil {
			zap.L().Error("could not stop server", zap.Error(err))
			return err
		}
		return err
	}

	return nil
}

// startHTTPServer запускает обычный HTTP сервер.
func (s *Server) startHTTPServer() error {
	s.server = &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}

// startHTTPServerTLS запускает HTTPS сервер.
// Для localhost и IP-адресов использует локальные сертификаты,
// для публичных доменов использует Let's Encrypt.
func (s *Server) startHTTPServerTLS() error {
	// Извлекаем хост из адреса (без порта)
	host := strings.Split(s.config.BaseURLAddress, ":")[0]
	isLocalhost := strings.Contains(s.config.BaseURLAddress, "localhost")
	isIP := net.ParseIP(host) != nil

	if isLocalhost || isIP {
		s.server = &http.Server{
			Addr:    s.config.ServerAddress,
			Handler: s.router,
		}
		return s.server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem")
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.BaseURLAddress),
	}
	s.server = &http.Server{
		Addr:      s.config.ServerAddress,
		TLSConfig: m.TLSConfig(),
		Handler:   s.router,
	}
	return s.server.ListenAndServeTLS("", "")
}
