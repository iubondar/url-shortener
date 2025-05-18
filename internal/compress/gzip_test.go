package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleWithGzipCompression демонстрирует базовое использование middleware для сжатия.
// Пример показывает, как middleware автоматически сжимает ответ сервера,
// если клиент поддерживает gzip-сжатие.
func ExampleWithGzipCompression() {
	// Создаем простой обработчик, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "application/json")
		io.WriteString(w, `{"message": "Hello, World!"}`)
	})

	// Оборачиваем обработчик в middleware для сжатия
	compressedHandler := WithGzipCompression(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(compressedHandler)
	defer server.Close()

	// Создаем запрос с поддержкой gzip
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "gzip")

	// Выполняем запрос
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	// Проверяем, что ответ сжат
	encoding := resp.Header.Get(contentEncoding)
	if encoding == "gzip" {
		// Распаковываем ответ
		reader, _ := gzip.NewReader(resp.Body)
		defer reader.Close()
		body, _ := io.ReadAll(reader)
		io.WriteString(io.Discard, string(body))
	}
}

// ExampleWithGzipCompression_compressedRequest демонстрирует обработку сжатых запросов.
// Пример показывает, как middleware автоматически распаковывает входящие запросы,
// если они сжаты с помощью gzip.
func ExampleWithGzipCompression_compressedRequest() {
	// Создаем простой обработчик, который читает тело запроса
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set(contentType, "application/json")
		io.WriteString(w, string(body))
	})

	// Оборачиваем обработчик в middleware для сжатия
	compressedHandler := WithGzipCompression(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(compressedHandler)
	defer server.Close()

	// Создаем сжатые данные
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(`{"message": "Compressed Request"}`))
	zw.Close()

	// Создаем запрос со сжатым телом
	req, _ := http.NewRequest("POST", server.URL, &buf)
	req.Header.Set(contentEncoding, "gzip")

	// Выполняем запрос
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	// Читаем ответ
	body, _ := io.ReadAll(resp.Body)
	io.WriteString(io.Discard, string(body))
}

func TestGzipCompression(t *testing.T) {
	requestBody := `
		<html><body><h1>Hello world!</h1></body></html>
	`

	successBody := `{
		"result": "https://127.0.0.1/abcdef11"
	}`

	withoutGzip := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "application/json")
		io.WriteString(w, successBody)
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

		defer resp.Body.Close()

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

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.JSONEq(t, successBody, string(b))
	})
}
