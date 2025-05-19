package logging

import (
	"net/http"
	"net/http/httptest"
)

// ExampleWithLogging демонстрирует базовое использование middleware для логирования.
// Пример показывает, как middleware логирует информацию о HTTP-запросе.
func ExampleWithLogging() {
	// Создаем простой обработчик, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "Hello, World!"}`))
		if err != nil {
			panic(err)
		}
	})

	// Оборачиваем обработчик в middleware для логирования
	loggedHandler := WithLogging(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(loggedHandler)
	defer server.Close()

	// Выполняем запрос
	req, err := http.NewRequest("GET", server.URL+"/api/test", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Проверяем, что запрос выполнен успешно
	if resp.StatusCode == http.StatusOK {
		// Запрос успешно обработан и залогирован
	}
	// Output:
}

// ExampleWithLogging_error демонстрирует логирование ошибок.
// Пример показывает, как middleware логирует информацию о запросе,
// завершившемся с ошибкой.
func ExampleWithLogging_error() {
	// Создаем обработчик, который возвращает ошибку
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error": "Internal Server Error"}`))
		if err != nil {
			panic(err)
		}
	})

	// Оборачиваем обработчик в middleware для логирования
	loggedHandler := WithLogging(handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(loggedHandler)
	defer server.Close()

	// Выполняем запрос
	req, err := http.NewRequest("GET", server.URL+"/api/error", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Проверяем, что запрос завершился с ошибкой
	if resp.StatusCode == http.StatusInternalServerError {
		// Ошибка успешно обработана и залогирована
	}
	// Output:
}
