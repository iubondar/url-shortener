package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/models"
	simple_storage "github.com/iubondar/url-shortener/internal/app/storage/simple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://haykot.dev/blog/til-testing-parametrized-urls-with-chi-router/
//
// withURLParam returns a pointer to a request object with the given URL params
// added to a new chi.Context object.
func withURLParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return req
}

// ExampleRetrieveURLHandler_RetrieveURL демонстрирует пример использования эндпоинта получения оригинального URL.
// Пример показывает, как получить оригинальный URL по сокращенному идентификатору.
func ExampleRetrieveURLHandler_RetrieveURL() {
	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = withURLParam(request, "id", "123")

	// Создаем репозиторий с тестовыми данными
	repo := &simple_storage.SimpleRepository{
		Records: []models.Record{
			{
				ShortURL:    "123",
				OriginalURL: "https://example.com",
				UserID:      uuid.New(),
			},
		},
	}

	// Инициализируем обработчик
	handler := NewRetrieveURLHandler(repo)

	// Вызываем обработчик
	w := httptest.NewRecorder()
	handler.RetrieveURL(w, request)

	// Получаем ответ
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Выводим статус ответа и заголовок Location
	fmt.Println(res.Status)
	fmt.Println(res.Header.Get("Location"))
	// Output:
	// 307 Temporary Redirect
	// https://example.com
}

func TestRetrieveURLHandler_RetrieveURL(t *testing.T) {
	userID := uuid.New()
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		method string
		id     string
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodGet,
			id:     "123",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name:   "Test POST method not allowed",
			method: http.MethodPost,
			id:     "123",
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			id:     "123",
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			id:     "123",
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "Test deleted URL",
			method: http.MethodGet,
			id:     "456",
			want: want{
				code:     http.StatusGone,
				location: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := simple_storage.SimpleRepository{
				Records: []models.Record{
					{
						ShortURL:    "123",
						OriginalURL: testURL,
						UserID:      userID,
					},
					{
						ShortURL:    "456",
						OriginalURL: "http://abc.com",
						UserID:      userID,
						IsDeleted:   true,
					},
				},
			}
			handler := NewRetrieveURLHandler(&repo)

			request := httptest.NewRequest(test.method, "/", nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			handler.RetrieveURL(w, withURLParam(request, "id", test.id))

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}()

			// проверяем код ответа, выходим если он ошибочный
			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// проверяем нужный заголовок
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestRetrieveURLHandler_WithNoIdParameter(t *testing.T) {
	repo := simple_storage.SimpleRepository{
		Records: []models.Record{
			{
				ShortURL:    "123",
				OriginalURL: testURL,
				UserID:      uuid.New(),
			},
		},
	}
	handler := NewRetrieveURLHandler(&repo)
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	// создаём новый Recorder
	w := httptest.NewRecorder()

	handler.RetrieveURL(w, request)

	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRetrieveURLHandler_WithNoURL(t *testing.T) {
	handler := NewRetrieveURLHandler(simple_storage.NewSimpleRepository())
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("id", "123")

	// создаём новый Recorder
	w := httptest.NewRecorder()

	handler.RetrieveURL(w, request)

	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
