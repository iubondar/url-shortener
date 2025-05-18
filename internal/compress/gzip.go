// Пакет compress предоставляет middleware для сжатия HTTP-трафика с использованием gzip.
// Поддерживает как сжатие ответов сервера, так и распаковку запросов от клиента.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressingContentTypes содержит список типов контента, которые должны сжиматься.
// По умолчанию сжимаются JSON и HTML.
var compressingContentTypes []string = []string{"application/json", "text/html"}

const (
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	contentType     = "Content-Type"
)

// shouldCompress проверяет, нужно ли сжимать контент указанного типа.
// Возвращает true, если тип контента входит в список compressingContentTypes.
func shouldCompress(ct string) bool {
	for _, contentType := range compressingContentTypes {
		if strings.Contains(ct, contentType) {
			return true
		}
	}
	return false
}

// gzipWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки.
type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newGzipWriter создает новый экземпляр gzipWriter.
// Оборачивает http.ResponseWriter для сжатия исходящего трафика.
func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	zw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &gzipWriter{
		w:  w,
		zw: zw,
	}
}

// Header возвращает HTTP-заголовки ответа.
func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в ответ, сжимая их если необходимо.
// Сжатие применяется только к определенным типам контента.
func (c *gzipWriter) Write(p []byte) (int, error) {
	ct := c.w.Header().Get(contentType)
	if shouldCompress(ct) {
		c.w.Header().Set(contentEncoding, "gzip")
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

// WriteHeader отправляет HTTP-заголовки ответа.
// Устанавливает заголовок Content-Encoding: gzip если контент должен быть сжат.
func (c *gzipWriter) WriteHeader(statusCode int) {
	ct := c.w.Header().Get(contentType)
	if shouldCompress(ct) && (statusCode < 300 || statusCode == http.StatusConflict) {
		c.w.Header().Set(contentEncoding, "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *gzipWriter) Close() error {
	ct := c.w.Header().Get(contentType)
	if shouldCompress(ct) {
		return c.zw.Close()
	}
	return nil
}

// gzipReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newGzipReader создает новый экземпляр gzipReader.
// Оборачивает io.ReadCloser для распаковки входящего трафика.
func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и распаковывает данные из запроса.
func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader и оригинальный io.ReadCloser.
func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// WithGzipCompression создает middleware для поддержки gzip-сжатия.
// Middleware автоматически сжимает ответы сервера и распаковывает запросы клиента,
// если они используют gzip-сжатие.
//
// Middleware проверяет заголовки Accept-Encoding и Content-Encoding
// для определения необходимости сжатия/распаковки.
func WithGzipCompression(h http.Handler) http.Handler {
	compressFn := func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get(acceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newGzipWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get(contentEncoding)
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(compressFn)
}
