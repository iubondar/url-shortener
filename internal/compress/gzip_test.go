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
