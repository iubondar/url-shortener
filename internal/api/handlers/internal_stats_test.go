package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStatsRetriever - мок для тестирования StatsRetriever
type MockStatsRetriever struct {
	stats models.Stats
	err   error
}

func (m *MockStatsRetriever) GetStats(ctx context.Context) (models.Stats, error) {
	return m.stats, m.err
}

// ExampleInternalStatsHandler_GetStats демонстрирует пример использования эндпоинта получения внутренней статистики.
// Пример показывает, как получить статистику о количестве сокращенных URL и пользователей.
func ExampleInternalStatsHandler_GetStats() {
	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

	// Создаем мок с тестовыми данными
	mockRetriever := &MockStatsRetriever{
		stats: models.Stats{
			URLsCount:  100,
			UsersCount: 25,
		},
	}

	// Инициализируем обработчик
	handler := NewInternalStatsHandler(mockRetriever)

	// Вызываем обработчик
	w := httptest.NewRecorder()
	handler.GetStats(w, request)

	// Получаем ответ
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Выводим статус ответа
	fmt.Println(res.Status)
	// Output: 200 OK
}

func TestInternalStatsHandler_GetStats(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		mockStats    models.Stats
		mockError    error
		wantCode     int
		wantStats    models.Stats
		wantErrorMsg string
	}{
		{
			name:   "Positive test - successful stats retrieval",
			method: http.MethodGet,
			mockStats: models.Stats{
				URLsCount:  150,
				UsersCount: 30,
			},
			mockError: nil,
			wantCode:  http.StatusOK,
			wantStats: models.Stats{
				URLsCount:  150,
				UsersCount: 30,
			},
		},
		{
			name:   "Zero stats test",
			method: http.MethodGet,
			mockStats: models.Stats{
				URLsCount:  0,
				UsersCount: 0,
			},
			mockError: nil,
			wantCode:  http.StatusOK,
			wantStats: models.Stats{
				URLsCount:  0,
				UsersCount: 0,
			},
		},
		{
			name:         "POST method not allowed",
			method:       http.MethodPost,
			mockStats:    models.Stats{},
			mockError:    nil,
			wantCode:     http.StatusMethodNotAllowed,
			wantErrorMsg: "Only GET requests are allowed!\n",
		},
		{
			name:         "PUT method not allowed",
			method:       http.MethodPut,
			mockStats:    models.Stats{},
			mockError:    nil,
			wantCode:     http.StatusMethodNotAllowed,
			wantErrorMsg: "Only GET requests are allowed!\n",
		},
		{
			name:         "DELETE method not allowed",
			method:       http.MethodDelete,
			mockStats:    models.Stats{},
			mockError:    nil,
			wantCode:     http.StatusMethodNotAllowed,
			wantErrorMsg: "Only GET requests are allowed!\n",
		},
		{
			name:         "StatsRetriever error",
			method:       http.MethodGet,
			mockStats:    models.Stats{},
			mockError:    fmt.Errorf("database connection failed"),
			wantCode:     http.StatusInternalServerError,
			wantErrorMsg: "database connection failed\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Создаем тестовый запрос
			request := httptest.NewRequest(test.method, "/api/internal/stats", nil)
			w := httptest.NewRecorder()

			// Создаем мок с тестовыми данными
			mockRetriever := &MockStatsRetriever{
				stats: test.mockStats,
				err:   test.mockError,
			}

			// Инициализируем обработчик
			handler := NewInternalStatsHandler(mockRetriever)

			// Вызываем обработчик
			handler.GetStats(w, request)

			// Получаем ответ
			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}()

			// Проверяем статус код
			require.Equal(t, test.wantCode, res.StatusCode)

			// Читаем тело ответа
			var buf bytes.Buffer
			_, err := buf.ReadFrom(res.Body)
			require.NoError(t, err)

			// Проверяем ответ в зависимости от ожидаемого статуса
			if test.wantCode == http.StatusOK {
				// Проверяем Content-Type для успешных ответов
				assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

				// Декодируем JSON ответ
				var responseStats models.Stats
				err = json.Unmarshal(buf.Bytes(), &responseStats)
				require.NoError(t, err)

				// Проверяем, что статистика соответствует ожидаемой
				assert.Equal(t, test.wantStats, responseStats)
			} else {
				// Проверяем сообщение об ошибке для неуспешных ответов
				assert.Equal(t, test.wantErrorMsg, buf.String())
			}
		})
	}
}

func TestNewInternalStatsHandler(t *testing.T) {
	// Создаем мок
	mockRetriever := &MockStatsRetriever{}

	// Создаем обработчик
	handler := NewInternalStatsHandler(mockRetriever)

	// Проверяем, что обработчик создан корректно
	assert.NotNil(t, handler)
	assert.Equal(t, mockRetriever, handler.statsRetriever)
}
