package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/models"
	simple_storage "github.com/iubondar/url-shortener/internal/app/storage/simple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleDeleteUrlsHandler_DeleteUserURLs демонстрирует пример использования эндпоинта удаления сокращенных ссылок.
// Пример показывает, как удалить сокращенные ссылки пользователя.
func ExampleDeleteUrlsHandler_DeleteUserURLs() {
	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte(`["123", "456"]`)))
	userID := uuid.New()
	authCookie, _ := auth.NewAuthCookie(userID)
	request.AddCookie(authCookie)

	w := httptest.NewRecorder()

	// Инициализируем репозиторий и обработчик
	repo := &simple_storage.SimpleRepository{}
	handler := NewDeleteUrlsHandler(repo)

	// Вызываем обработчик
	handler.DeleteUserURLs(w, request)

	// Получаем ответ
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Выводим статус ответа
	fmt.Println(res.Status)
	// Output: 202 Accepted
}

func TestDeleteUrlsHandler_DeleteUserURLs(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name        string
		method      string
		records     []models.Record
		body        string
		userID      uuid.UUID
		wantCode    int
		wantRecords []models.Record
	}{
		{
			name:   "Positive test",
			method: http.MethodDelete,
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
				},
			},
			body:     "[\"123\"]",
			userID:   userID,
			wantCode: http.StatusAccepted,
			wantRecords: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: "http://example.com",
					UserID:      userID,
					IsDeleted:   true,
				},
			},
		},
		{
			name:        "Invalid JSON",
			method:      http.MethodDelete,
			records:     []models.Record{},
			body:        "[\"123\", 456,]",
			userID:      userID,
			wantCode:    http.StatusBadRequest,
			wantRecords: []models.Record{},
		},
		{
			name:        "GET method not allowed",
			method:      http.MethodGet,
			records:     []models.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []models.Record{},
		},
		{
			name:        "POST method not allowed",
			method:      http.MethodPost,
			records:     []models.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []models.Record{},
		},
		{
			name:        "PUT method not allowed",
			method:      http.MethodPut,
			records:     []models.Record{},
			body:        "[\"123\"]",
			userID:      userID,
			wantCode:    http.StatusMethodNotAllowed,
			wantRecords: []models.Record{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := httptest.NewRequest(test.method, "/api/user/urls", bytes.NewReader([]byte(test.body)))
			authCookie, err := auth.NewAuthCookie(test.userID)
			require.NoError(t, err)
			request.AddCookie(authCookie)

			w := httptest.NewRecorder()
			repo := simple_storage.SimpleRepository{
				Records: test.records,
			}
			handler := NewDeleteUrlsHandler(&repo)

			handler.DeleteUserURLs(w, request)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}()

			assert.Equal(t, test.wantCode, res.StatusCode)
			assert.ElementsMatch(t, test.wantRecords, repo.Records)
		})
	}
}
