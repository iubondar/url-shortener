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

const testURL string = "https://practicum.yandex.ru"

func TestShortenHandler_Shorten(t *testing.T) {
	type fields struct {
		urlsToIds map[string]string
		idsToURLs map[string]string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		body   string
		fields fields
		want   want
	}{
		{
			name:   "Positive test",
			method: http.MethodPost,
			body:   "{\"url\": \"" + testURL + "\"}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusCreated,
				response:    `http://127.0.0.1`,
				contentType: "application/json",
			},
		},
		{
			name:   "Existed record test",
			method: http.MethodPost,
			body:   "{\"url\": \"" + testURL + "\"}",
			fields: fields{
				urlsToIds: map[string]string{testURL: "123"},
				idsToURLs: map[string]string{"123": testURL},
			},
			want: want{
				code:        http.StatusConflict,
				response:    `http://127.0.0.1`,
				contentType: "application/json",
			},
		},
		{
			name:   "Test invalid json in request",
			method: http.MethodPost,
			body:   "{url: " + testURL + "}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test invalid URL",
			method: http.MethodPost,
			body:   "{\"url\": \"htps/practicum.yandex.ru\"}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test GET method not allowed",
			method: http.MethodGet,
			body:   "{\"url\": \"" + testURL + "\"}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test PUT method not allowed",
			method: http.MethodPut,
			body:   "{\"url\": \"" + testURL + "\"}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
		{
			name:   "Test DELETE method not allowed",
			method: http.MethodDelete,
			body:   "{\"url\": \"" + testURL + "\"}",
			fields: fields{
				urlsToIds: map[string]string{},
				idsToURLs: map[string]string{},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    `http://127.0.0.1`,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", bytes.NewReader([]byte(test.body)))
			// создаём новый Recorder
			w := httptest.NewRecorder()
			repo := storage.SimpleRepository{
				UrlsToIds: test.fields.urlsToIds,
				IdsToURLs: test.fields.idsToURLs,
			}
			handler := NewShortenHandler(repo, "127.0.0.1")
			handler.Shorten(w, request)

			res := w.Result()

			// проверяем код ответа, выходим если он ошибочный
			require.Equal(t, test.want.code, res.StatusCode)
			if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
				return
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			var buf bytes.Buffer
			_, err := buf.ReadFrom(res.Body)
			require.NoError(t, err)

			var out ShortenOut
			err = json.Unmarshal(buf.Bytes(), &out)
			require.NoError(t, err)

			id, err := repo.RetrieveID(testURL)
			require.NoError(t, err)
			assert.Equal(t, test.want.response+"/"+id, out.Result)
		})
	}
}
