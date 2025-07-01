package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	request.Header.Set("X-Real-IP", "192.168.1.1")

	// Создаем мок с тестовыми данными
	mockRetriever := &MockStatsRetriever{
		stats: models.Stats{
			URLsCount:  100,
			UsersCount: 25,
		},
	}

	// Инициализируем обработчик
	handler := NewInternalStatsHandler(mockRetriever, "192.168.1.0/24")

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
		name          string
		method        string
		trustedSubnet string
		realIP        string
		mockStats     models.Stats
		mockError     error
		wantCode      int
		wantStats     models.Stats
		wantErrorMsg  string
	}{
		{
			name:          "Positive test - successful stats retrieval",
			method:        http.MethodGet,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
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
			name:          "Zero stats test",
			method:        http.MethodGet,
			trustedSubnet: "10.0.0.0/8",
			realIP:        "10.0.0.1",
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
			name:          "POST method not allowed",
			method:        http.MethodPost,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusMethodNotAllowed,
			wantErrorMsg:  "Only GET requests are allowed!\n",
		},
		{
			name:          "PUT method not allowed",
			method:        http.MethodPut,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusMethodNotAllowed,
			wantErrorMsg:  "Only GET requests are allowed!\n",
		},
		{
			name:          "DELETE method not allowed",
			method:        http.MethodDelete,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusMethodNotAllowed,
			wantErrorMsg:  "Only GET requests are allowed!\n",
		},
		{
			name:          "StatsRetriever error",
			method:        http.MethodGet,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			mockStats:     models.Stats{},
			mockError:     fmt.Errorf("database connection failed"),
			wantCode:      http.StatusInternalServerError,
			wantErrorMsg:  "database connection failed\n",
		},
		{
			name:          "Empty trusted subnet - access denied",
			method:        http.MethodGet,
			trustedSubnet: "",
			realIP:        "192.168.1.100",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusForbidden,
			wantErrorMsg:  "Trusted subnet is not set - access denied\n",
		},
		{
			name:          "Missing X-Real-IP header",
			method:        http.MethodGet,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusForbidden,
			wantErrorMsg:  "X-Real-IP header is required\n",
		},
		{
			name:          "Invalid X-Real-IP header",
			method:        http.MethodGet,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "invalid-ip",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusForbidden,
			wantErrorMsg:  "Invalid X-Real-IP header\n",
		},
		{
			name:          "IP not in trusted subnet",
			method:        http.MethodGet,
			trustedSubnet: "192.168.1.0/24",
			realIP:        "10.0.0.1",
			mockStats:     models.Stats{},
			mockError:     nil,
			wantCode:      http.StatusForbidden,
			wantErrorMsg:  "Access denied - IP not in trusted subnet\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Создаем тестовый запрос
			request := httptest.NewRequest(test.method, "/api/internal/stats", nil)
			if test.realIP != "" {
				request.Header.Set("X-Real-IP", test.realIP)
			}
			w := httptest.NewRecorder()

			// Создаем мок с тестовыми данными
			mockRetriever := &MockStatsRetriever{
				stats: test.mockStats,
				err:   test.mockError,
			}

			// Инициализируем обработчик
			handler := NewInternalStatsHandler(mockRetriever, test.trustedSubnet)

			// Вызываем обработчик
			handler.GetStats(w, request)

			// Получаем ответ
			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}()

			// Проверяем код ответа
			assert.Equal(t, test.wantCode, res.StatusCode)

			// Если ожидаем ошибку, проверяем сообщение
			if test.wantErrorMsg != "" {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, test.wantErrorMsg, string(body))
				return
			}

			// Если успешный ответ, проверяем статистику
			if res.StatusCode == http.StatusOK {
				assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

				var stats models.Stats
				err := json.NewDecoder(res.Body).Decode(&stats)
				require.NoError(t, err)

				assert.Equal(t, test.wantStats.URLsCount, stats.URLsCount)
				assert.Equal(t, test.wantStats.UsersCount, stats.UsersCount)
			}
		})
	}
}

func TestNewInternalStatsHandler(t *testing.T) {
	mockRetriever := &MockStatsRetriever{}
	trustedSubnet := "192.168.1.0/24"

	handler := NewInternalStatsHandler(mockRetriever, trustedSubnet)

	assert.Equal(t, mockRetriever, handler.statsRetriever)
	assert.Equal(t, trustedSubnet, handler.trustedSubnet)
}

// TestInternalStatsHandler_GetStats_WriteResponseError тестирует обработку ошибки записи ответа
func TestInternalStatsHandler_GetStats_WriteResponseError(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	request.Header.Set("X-Real-IP", "192.168.1.100")

	w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder()}

	mockRetriever := &MockStatsRetriever{
		stats: models.Stats{
			URLsCount:  100,
			UsersCount: 25,
		},
	}

	handler := NewInternalStatsHandler(mockRetriever, "192.168.1.0/24")

	handler.GetStats(w, request)

	// Проверяем, что обработчик корректно обработал ошибку записи
	assert.True(t, w.writeCalled)
}

// TestInternalStatsHandler_GetStats_InvalidTrustedSubnet тестирует обработку невалидной подсети
func TestInternalStatsHandler_GetStats_InvalidTrustedSubnet(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	request.Header.Set("X-Real-IP", "192.168.1.100")

	w := httptest.NewRecorder()

	mockRetriever := &MockStatsRetriever{}

	handler := NewInternalStatsHandler(mockRetriever, "invalid-subnet")

	handler.GetStats(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
