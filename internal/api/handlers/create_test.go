package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/models"
	simple_storage "github.com/iubondar/url-shortener/internal/app/storage/simple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleCreateIDHandler_CreateID демонстрирует пример использования эндпоинта создания сокращенной ссылки.
// Пример показывает, как создать сокращенную ссылку для длинного URL.
func ExampleCreateIDHandler_CreateID() {
	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
	w := httptest.NewRecorder()

	// Инициализируем репозиторий и обработчик
	repo := &simple_storage.SimpleRepository{}
	handler := NewCreateIDHandler(repo, "127.0.0.1")

	// Вызываем обработчик
	handler.CreateID(w, request)

	// Получаем ответ
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	// Выводим статус ответа
	fmt.Println(res.Status)
	// Output: 201 Created
}

func TestCreateIDHandler_CreateID(t *testing.T) {
	userID := uuid.New()
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		url     string
		records []models.Record
		want    want
	}{
		{
			name:    "Positive test",
			method:  http.MethodPost,
			url:     "https://practicum.yandex.ru/",
			records: []models.Record{},
			want: want{
				code:        http.StatusCreated,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:   "Existed record test",
			method: http.MethodPost,
			url:    testURL,
			records: []models.Record{
				{
					ShortURL:    "123",
					OriginalURL: testURL,
					UserID:      userID,
				},
			},
			want: want{
				code:        http.StatusConflict,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:    "Test invalid URL",
			method:  http.MethodPost,
			url:     "https/practicum.yandex.ru/",
			records: []models.Record{},
			want: want{
				code:        http.StatusBadRequest,
				response:    `http://127.0.0.1`,
				contentType: "text/plain",
			},
		},
		{
			name:    "Test GET method not allowed",
			method:  http.MethodGet,
			url:     "https://practicum.yandex.ru/",
			records: []models.Record{},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:    "Test PUT method not allowed",
			method:  http.MethodPut,
			url:     "https://practicum.yandex.ru/",
			records: []models.Record{},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:    "Test DELETE method not allowed",
			method:  http.MethodDelete,
			url:     "https://practicum.yandex.ru/",
			records: []models.Record{},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", bytes.NewReader([]byte(test.url)))
			// создаём новый Recorder
			w := httptest.NewRecorder()
			repo := simple_storage.SimpleRepository{
				Records: test.records,
			}
			handler := NewCreateIDHandler(&repo, "127.0.0.1")
			handler.CreateID(w, request)

			res := w.Result()
			// проверяем код ответа, выходим если он ошибочный
			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// получаем и проверяем тело запроса
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Error closing response body: %v", err)
				}
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			id, err := repo.RetrieveID(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.want.response+"/"+id, string(resBody))
		})
	}
}

// TestCreateIDHandler_CreateID_ReadBodyError тестирует обработку ошибки чтения тела запроса
func TestCreateIDHandler_CreateID_ReadBodyError(t *testing.T) {
	// Создаем запрос с телом, которое нельзя прочитать
	request := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
	w := httptest.NewRecorder()

	repo := simple_storage.SimpleRepository{}
	handler := NewCreateIDHandler(&repo, "127.0.0.1")

	handler.CreateID(w, request)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

// errorReader - io.Reader, который всегда возвращает ошибку
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

// TestCreateIDHandler_CreateID_WriteResponseError тестирует обработку ошибки записи ответа
func TestCreateIDHandler_CreateID_WriteResponseError(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("https://example.com")))
	w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder()}

	repo := simple_storage.SimpleRepository{}
	handler := NewCreateIDHandler(&repo, "127.0.0.1")

	handler.CreateID(w, request)

	// Проверяем, что обработчик корректно обработал ошибку записи
	assert.True(t, w.writeCalled)
}

// errorResponseWriter - http.ResponseWriter, который возвращает ошибку при записи
type errorResponseWriter struct {
	http.ResponseWriter
	writeCalled bool
}

func (e *errorResponseWriter) Write(data []byte) (int, error) {
	e.writeCalled = true
	return 0, fmt.Errorf("write error")
}
