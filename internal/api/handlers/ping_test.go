package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/iubondar/url-shortener/internal/app/storage/mocks"
	"github.com/stretchr/testify/assert"
)

// ExamplePingHandler_Ping демонстрирует пример использования эндпоинта проверки доступности сервиса.
// Пример показывает, как проверить работоспособность сервера.
func ExamplePingHandler_Ping() {
	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// Создаем мок для проверки статуса
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockChecker := mocks.NewMockStatusChecker(ctrl)
	mockChecker.EXPECT().CheckStatus(gomock.Any()).Return(nil)

	// Инициализируем обработчик
	handler := NewPingHandler(mockChecker)

	// Вызываем обработчик
	handler.Ping(w, request)

	// Получаем ответ
	res := w.Result()
	defer res.Body.Close()

	// Выводим статус ответа
	fmt.Println(res.Status)
	// Output: 200 OK
}

func TestPingHandler_Ping(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		setErr   error
		wantCode int
	}{
		{
			name:     "Positive test",
			method:   http.MethodGet,
			setErr:   nil,
			wantCode: http.StatusOK,
		},
		{
			name:     "Test POST method not allowed",
			method:   http.MethodPost,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Test PUT method not allowed",
			method:   http.MethodPut,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Test DELETE method not allowed",
			method:   http.MethodDelete,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Check error test",
			method:   http.MethodGet,
			setErr:   errors.New("Status is not ok"),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockStatusChecker(ctrl)

			m.EXPECT().CheckStatus(gomock.Any()).Return(tt.setErr).AnyTimes()

			request := httptest.NewRequest(tt.method, "/ping", nil)

			w := httptest.NewRecorder()

			handler := PingHandler{checker: m}
			handler.Ping(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}
