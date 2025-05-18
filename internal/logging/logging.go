// Пакет logging предоставляет middleware для логирования HTTP-запросов.
// Использует zap для структурированного логирования с информацией о запросе,
// ответе и времени выполнения.
package logging

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData хранит информацию об ответе сервера
	responseData struct {
		status int // HTTP-код ответа
		size   int // размер тела ответа в байтах
	}

	// loggingResponseWriter реализует интерфейс http.ResponseWriter
	// и перехватывает информацию о записываемом ответе.
	// Встраивает оригинальный http.ResponseWriter для сохранения его функциональности.
	loggingResponseWriter struct {
		http.ResponseWriter               // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData // данные об ответе
	}
)

// Write перехватывает запись тела ответа и сохраняет его размер.
// Реализует метод интерфейса http.ResponseWriter.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader перехватывает установку кода статуса ответа.
// Реализует метод интерфейса http.ResponseWriter.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging создает middleware для логирования HTTP-запросов.
// Логирует следующую информацию о каждом запросе:
// - URI запроса
// - HTTP метод
// - Код статуса ответа
// - Время выполнения запроса
// - Размер ответа в байтах
//
// Использует zap для структурированного логирования в режиме разработки.
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatalf("can't initialize zap logger: %v", err)
		}
		defer logger.Sync()

		sugar := *logger.Sugar()

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return http.HandlerFunc(logFn)
}
