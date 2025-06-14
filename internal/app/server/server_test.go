package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iubondar/url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	// Инициализируем логгер для тестов
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	cfg := config.Config{
		ServerAddress:  ":8080",
		BaseURLAddress: "http://localhost:8080",
		EnableHTTPS:    false,
	}

	router := http.NewServeMux()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := New(cfg, router)
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, router, server.router)
}

func TestServerStartAndShutdown(t *testing.T) {
	// Инициализируем логгер для тестов
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	tests := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "HTTP server",
			config: config.Config{
				ServerAddress:  ":0", // Используем порт 0 для автоматического выбора свободного порта
				BaseURLAddress: "http://localhost",
				EnableHTTPS:    false,
			},
			expectError: false,
		},
		{
			name: "HTTPS server with localhost",
			config: config.Config{
				ServerAddress:  ":0",
				BaseURLAddress: "https://localhost",
				EnableHTTPS:    true,
			},
			expectError: true, // Ожидаем ошибку, так как нет сертификатов
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := http.NewServeMux()
			router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server := New(tt.config, router)

			// Запускаем сервер в отдельной горутине
			errChan := make(chan error, 1)
			go func() {
				err := server.Start()
				// Игнорируем ошибку "Server closed" для HTTP сервера
				if err != nil && !strings.Contains(err.Error(), "Server closed") {
					errChan <- err
				} else {
					errChan <- nil
				}
			}()

			// Даем серверу время на запуск
			time.Sleep(100 * time.Millisecond)

			// Проверяем, что сервер запущен
			if !tt.expectError {
				// Делаем тестовый запрос
				client := &http.Client{
					Timeout: time.Second,
				}
				resp, err := client.Get("http://localhost" + tt.config.ServerAddress + "/test")
				if err == nil {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
					resp.Body.Close()
				}
			}

			// Выполняем graceful shutdown
			err := server.Shutdown()
			assert.NoError(t, err, "shutdown should not return error")

			// Проверяем ошибку запуска сервера
			select {
			case err := <-errChan:
				if tt.expectError {
					assert.Error(t, err, "HTTPS server should return error due to missing certificates")
					assert.True(t, strings.Contains(err.Error(), "cert.pem") ||
						strings.Contains(err.Error(), "no such file"),
						"error should be about missing certificates")
				} else {
					assert.NoError(t, err, "HTTP server should not return error")
				}
			case <-time.After(time.Second):
				t.Error("server did not return error in time")
			}
		})
	}
}

func TestServerHandler(t *testing.T) {
	// Инициализируем логгер для тестов
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	router := http.NewServeMux()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	cfg := config.Config{
		ServerAddress:  ":0",
		BaseURLAddress: "http://localhost",
		EnableHTTPS:    false,
	}

	server := New(cfg, router)

	// Создаем тестовый HTTP запрос
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик напрямую
	server.router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}
