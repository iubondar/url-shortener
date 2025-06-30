package testhelpers

import (
	"fmt"
	"net/http"
)

type ErrorResponseWriter struct {
	http.ResponseWriter
	WriteCalled bool
}

func (e *ErrorResponseWriter) Write(data []byte) (int, error) {
	e.WriteCalled = true
	return 0, fmt.Errorf("write error")
}
