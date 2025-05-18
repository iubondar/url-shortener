package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
)

// ExampleWithGzipCompression демонстрирует базовое использование middleware для сжатия.
// Пример показывает, как middleware автоматически сжимает ответ сервера,
// если клиент поддерживает gzip-сжатие.
func ExampleWithGzipCompression() {
	// Создаем простой обработчик, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "application/json")
		_, err := io.WriteString(w, `{"message": "Hello, World!"}`)
		if err != nil {
			panic(err)
		}
	})

	// Оборачиваем обработчик в middleware для сжатия
	compressedHandler := WithGzipCompression(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(compressedHandler)
	defer server.Close()

	// Создаем запрос с поддержкой gzip
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set(acceptEncoding, "gzip")

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Проверяем, что ответ сжат
	encoding := resp.Header.Get(contentEncoding)
	if encoding == "gzip" {
		// Распаковываем ответ
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		_, err = io.Discard.Write(body)
		if err != nil {
			panic(err)
		}
	}
	// Output:
}

// ExampleWithGzipCompression_compressedRequest демонстрирует обработку сжатых запросов.
// Пример показывает, как middleware автоматически распаковывает входящие запросы,
// если они сжаты с помощью gzip.
func ExampleWithGzipCompression_compressedRequest() {
	// Создаем простой обработчик, который читает тело запроса
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		w.Header().Set(contentType, "application/json")
		_, err = w.Write(body)
		if err != nil {
			panic(err)
		}
	})

	// Оборачиваем обработчик в middleware для сжатия
	compressedHandler := WithGzipCompression(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(compressedHandler)
	defer server.Close()

	// Создаем сжатые данные
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(`{"message": "Compressed Request"}`))
	if err != nil {
		panic(err)
	}
	err = zw.Close()
	if err != nil {
		panic(err)
	}

	// Создаем запрос со сжатым телом
	req, err := http.NewRequest("POST", server.URL, &buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set(contentEncoding, "gzip")

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	_, err = io.Discard.Write(body)
	if err != nil {
		panic(err)
	}
	// Output:
}
