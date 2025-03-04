package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/iubondar/url-shortener/internal/app/storage/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler_Ping(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		setErr   error
		wantCode int
	}{
		{
			name:     "Positive test",
			method:   http.MethodGet,
			setErr:   nil,
			wantCode: http.StatusOK,
		},
		{
			name:     "Test POST method not allowed",
			method:   http.MethodPost,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Test PUT method not allowed",
			method:   http.MethodPut,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Test DELETE method not allowed",
			method:   http.MethodDelete,
			setErr:   nil,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "Check error test",
			method:   http.MethodGet,
			setErr:   errors.New("Status is not ok"),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockStatusChecker(ctrl)

			m.EXPECT().CheckStatus(gomock.Any()).Return(tt.setErr).AnyTimes()

			request := httptest.NewRequest(tt.method, "/ping", nil)

			w := httptest.NewRecorder()

			handler := PingHandler{checker: m}
			handler.Ping(w, request)

			res := w.Result()

			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}
