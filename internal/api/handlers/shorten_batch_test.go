package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenBatchHandler_ShortenBatch(t *testing.T) {
	type fields struct {
		urlsToIds map[string]string
		idsToURLs map[string]string
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusOK,
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
				urlsToIds: map[string]string{"http://practicum.yandex.ru": "098"},
				idsToURLs: map[string]string{"098": "http://practicum.yandex.ru"},
			},
			want: want{
				code:        http.StatusOK,
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
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
				UrlsToIds: test.fields.urlsToIds,
				IdsToURLs: test.fields.idsToURLs,
			}
			handler := NewShortenBatchHandler(repo, "127.0.0.1")
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
