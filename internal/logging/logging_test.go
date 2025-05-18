package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ExampleWithLogging демонстрирует базовое использование middleware для логирования.
// Пример показывает, как middleware логирует информацию о HTTP-запросе.
func ExampleWithLogging() {
	// Создаем простой обработчик, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, World!"}`))
	})

	// Оборачиваем обработчик в middleware для логирования
	loggedHandler := WithLogging(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(loggedHandler)
	defer server.Close()

	// Выполняем запрос
	req, _ := http.NewRequest("GET", server.URL+"/api/test", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	// Проверяем, что запрос выполнен успешно
	if resp.StatusCode == http.StatusOK {
		// Запрос успешно обработан и залогирован
	}
}

// ExampleWithLogging_error демонстрирует логирование ошибок.
// Пример показывает, как middleware логирует информацию о запросе,
// завершившемся с ошибкой.
func ExampleWithLogging_error() {
	// Создаем обработчик, который возвращает ошибку
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
	})

	// Оборачиваем обработчик в middleware для логирования
	loggedHandler := WithLogging(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(loggedHandler)
	defer server.Close()

	// Выполняем запрос
	req, _ := http.NewRequest("GET", server.URL+"/api/error", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	// Проверяем, что запрос завершился с ошибкой
	if resp.StatusCode == http.StatusInternalServerError {
		// Ошибка успешно обработана и залогирована
	}
}

func TestWithLogging(t *testing.T) {
	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Оборачиваем в middleware
	loggedHandler := WithLogging(handler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	loggedHandler.ServeHTTP(w, req)

	// Проверяем, что запрос обработан успешно
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}
