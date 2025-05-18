package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleShortenBatchHandler_ShortenBatch демонстрирует пример использования эндпоинта пакетного создания сокращенных URL.
// Пример показывает, как создать несколько сокращенных URL за один запрос.
func ExampleShortenBatchHandler_ShortenBatch() {
	// Создаем тестовые данные
	input := []ShortenBatchIn{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}
	jsonIn, _ := json.Marshal(input)

	// Создаем тестовый HTTP запрос
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonIn))
	w := httptest.NewRecorder()

	// Инициализируем репозиторий и обработчик
	repo := &storage.SimpleRepository{}
	handler := NewShortenBatchHandler(repo, "127.0.0.1")

	// Вызываем обработчик
	handler.ShortenBatch(w, request)

	// Получаем ответ
	res := w.Result()
	defer res.Body.Close()

	// Выводим статус ответа
	fmt.Println(res.Status)
	// Output: 201 Created
}

func TestShortenBatchHandler_ShortenBatch(t *testing.T) {
	userID := uuid.New()
	type fields struct {
		records []storage.Record
	}
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		in     []ShortenBatchIn
		fields fields
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodPost,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "http://ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:   "Some existed records test",
			method: http.MethodPost,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "http://ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{
					{
						ShortURL:    "098",
						OriginalURL: "http://practicum.yandex.ru",
						UserID:      userID,
					},
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:   "Test invalid URL",
			method: http.MethodPost,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name:   "Test GET method not allowed",
			method: http.MethodGet,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "http://ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "http://ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			in: []ShortenBatchIn{
				{CorrelationID: "123", OriginalURL: "http://yandex.ru"},
				{CorrelationID: "123", OriginalURL: "http://ya.ru"},
				{CorrelationID: "123", OriginalURL: "http://practicum.yandex.ru"},
			},
			fields: fields{
				records: []storage.Record{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonIn, err := json.Marshal(test.in)
			require.NoError(t, err)
			request := httptest.NewRequest(test.method, "/shorten/batch", bytes.NewReader(jsonIn))
			// создаём новый Recorder
			w := httptest.NewRecorder()
			repo := storage.SimpleRepository{
				Records: test.fields.records,
			}
			handler := NewShortenBatchHandler(&repo, "127.0.0.1")
			handler.ShortenBatch(w, request)

			res := w.Result()

			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			var buf bytes.Buffer
			_, err = buf.ReadFrom(res.Body)
			require.NoError(t, err)

			var out []ShortenBatchOut
			err = json.Unmarshal(buf.Bytes(), &out)
			require.NoError(t, err)

			for i, elem := range out {
				id, err := repo.RetrieveID(test.in[i].OriginalURL)
				require.NoError(t, err)
				assert.Equal(t, "http://127.0.0.1/"+id, elem.ShortURL)
				assert.Equal(t, test.in[i].CorrelationID, elem.CorrelationID)
			}
		})
	}
}
