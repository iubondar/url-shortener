package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {
	requestBody := `
		<html><body><h1>Hello world!</h1></body></html>
	`

	successBody := `{
		"result": "https://127.0.0.1/abcdef11"
	}`

	withoutGzip := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "application/json")
		_, err := io.WriteString(w, successBody)
		if err != nil {
			panic(err)
		}
	})
	handler := WithGzipCompression(withoutGzip)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set(contentEncoding, "gzip")
		r.Header.Set(acceptEncoding, "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Error closing response body: %v", err)
			}
		}()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set(contentType, "text/html")
		r.Header.Set(acceptEncoding, "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Error closing response body: %v", err)
			}
		}()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer func() {
			if err := zr.Close(); err != nil {
				t.Errorf("Error closing gzip reader: %v", err)
			}
		}()

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.JSONEq(t, successBody, string(b))
	})
}

func BenchmarkGzipCompression(b *testing.B) {
	// Тестовые данные разной длины и типов
	testCases := []struct {
		name        string
		contentType string
		content     string
	}{
		{
			name:        "small_json",
			contentType: "application/json",
			content:     `{"message": "Hello, World!"}`,
		},
		{
			name:        "medium_json",
			contentType: "application/json",
			content:     `{"items": [` + strings.Repeat(`{"id": 1, "name": "test"},`, 100) + `]}`,
		},
		{
			name:        "small_html",
			contentType: "text/html",
			content:     `<html><body><h1>Hello world!</h1></body></html>`,
		},
		{
			name:        "medium_html",
			contentType: "text/html",
			content:     `<html><body>` + strings.Repeat(`<p>Test paragraph</p>`, 100) + `</body></html>`,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, r.Header.Get(contentType))
		_, err := io.WriteString(w, r.Header.Get("X-Test-Content"))
		if err != nil {
			b.Fatal(err)
		}
	})

	compressedHandler := WithGzipCompression(handler)
	srv := httptest.NewServer(compressedHandler)
	defer srv.Close()

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			req, err := http.NewRequest("GET", srv.URL, nil)
			if err != nil {
				b.Fatal(err)
			}
			req.Header.Set(acceptEncoding, "gzip")
			req.Header.Set(contentType, tc.contentType)
			req.Header.Set("X-Test-Content", tc.content)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				_, err = io.ReadAll(resp.Body)
				if err != nil {
					b.Fatal(err)
				}
				if err := resp.Body.Close(); err != nil {
					b.Errorf("Error closing response body: %v", err)
				}
			}
		})
	}
}

// TestShouldCompress тестирует функцию shouldCompress
func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		should      bool
	}{
		{
			name:        "JSON content type",
			contentType: "application/json",
			should:      true,
		},
		{
			name:        "HTML content type",
			contentType: "text/html",
			should:      true,
		},
		{
			name:        "JSON with charset",
			contentType: "application/json; charset=utf-8",
			should:      true,
		},
		{
			name:        "HTML with charset",
			contentType: "text/html; charset=utf-8",
			should:      true,
		},
		{
			name:        "Plain text",
			contentType: "text/plain",
			should:      false,
		},
		{
			name:        "Empty content type",
			contentType: "",
			should:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldCompress(tt.contentType)
			assert.Equal(t, tt.should, result)
		})
	}
}

// TestGzipWriter_WriteHeader тестирует WriteHeader метод gzipWriter
func TestGzipWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который должен сжиматься
	gw.Header().Set(contentType, "application/json")
	gw.WriteHeader(http.StatusOK)

	assert.Equal(t, "gzip", w.Header().Get(contentEncoding))
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGzipWriter_WriteHeader_NonCompressible тестирует WriteHeader для несжимаемого контента
func TestGzipWriter_WriteHeader_NonCompressible(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который не должен сжиматься
	gw.Header().Set(contentType, "text/plain")
	gw.WriteHeader(http.StatusOK)

	assert.Equal(t, "", w.Header().Get(contentEncoding))
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGzipWriter_WriteHeader_ErrorStatus тестирует WriteHeader для статусов ошибок
func TestGzipWriter_WriteHeader_ErrorStatus(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который должен сжиматься
	gw.Header().Set(contentType, "application/json")
	gw.WriteHeader(http.StatusInternalServerError)

	// Для статусов ошибок сжатие не применяется
	assert.Equal(t, "", w.Header().Get(contentEncoding))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestGzipWriter_WriteHeader_ConflictStatus тестирует WriteHeader для статуса конфликта
func TestGzipWriter_WriteHeader_ConflictStatus(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который должен сжиматься
	gw.Header().Set(contentType, "application/json")
	gw.WriteHeader(http.StatusConflict)

	// Для статуса конфликта сжатие применяется
	assert.Equal(t, "gzip", w.Header().Get(contentEncoding))
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestGzipWriter_Close тестирует Close метод gzipWriter
func TestGzipWriter_Close(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который должен сжиматься
	gw.Header().Set(contentType, "application/json")

	err := gw.Close()
	assert.NoError(t, err)
}

// TestGzipWriter_Close_NonCompressible тестирует Close для несжимаемого контента
func TestGzipWriter_Close_NonCompressible(t *testing.T) {
	w := httptest.NewRecorder()
	gw := newGzipWriter(w)

	// Устанавливаем content type, который не должен сжиматься
	gw.Header().Set(contentType, "text/plain")

	err := gw.Close()
	assert.NoError(t, err)
}

// TestGzipReader_Close_Error тестирует Close метод gzipReader с ошибкой
func TestGzipReader_Close_Error(t *testing.T) {
	// Создаем gzipReader с невалидным reader
	gr := &gzipReader{
		r:  &errorReadCloser{},
		zr: nil, // nil reader вызовет ошибку при закрытии
	}

	err := gr.Close()
	assert.Error(t, err)
}

// errorReadCloser - io.ReadCloser, который возвращает ошибку при закрытии
type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (e *errorReadCloser) Close() error {
	return assert.AnError
}

// TestWithGzipCompression_InvalidGzipRequest тестирует обработку невалидного gzip запроса
func TestWithGzipCompression_InvalidGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	compressedHandler := WithGzipCompression(handler)
	srv := httptest.NewServer(compressedHandler)
	defer srv.Close()

	// Создаем запрос с невалидным gzip контентом
	req, err := http.NewRequest("POST", srv.URL, strings.NewReader("invalid gzip content"))
	require.NoError(t, err)
	req.Header.Set(contentEncoding, "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestWithGzipCompression_NoCompression тестирует обработку запроса без сжатия
func TestWithGzipCompression_NoCompression(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	compressedHandler := WithGzipCompression(handler)
	srv := httptest.NewServer(compressedHandler)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(acceptEncoding, "")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get(contentEncoding))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "test response", string(body))
}
