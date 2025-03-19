package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage"
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := storage.SimpleRepository{
				Records: []storage.Record{
					{
						ShortURL:    "123",
						OriginalURL: testURL,
						UserID:      userID,
					},
				},
			}
			handler := NewRetrieveURLHandler(&repo)

			request := httptest.NewRequest(test.method, "/", nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			handler.RetrieveURL(w, withURLParam(request, "id", test.id))

			res := w.Result()
			defer res.Body.Close()

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
	repo := storage.SimpleRepository{
		Records: []storage.Record{
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
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRetrieveURLHandler_WithNoURL(t *testing.T) {
	handler := NewRetrieveURLHandler(storage.NewSimpleRepository())
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("id", "123")

	// создаём новый Recorder
	w := httptest.NewRecorder()

	handler.RetrieveURL(w, request)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}
