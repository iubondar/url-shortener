package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogging(t *testing.T) {
	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test response"))
		if err != nil {
			t.Fatal(err)
		}
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
