package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserIDFromAuthCookieOrSetNew(t *testing.T) {
	t.Run("Get user ID from existing cookie", func(t *testing.T) {
		userID := uuid.New()
		request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		authCookie, err := NewAuthCookie(userID)
		require.NoError(t, err)
		request.AddCookie(authCookie)
		w := httptest.NewRecorder()

		gotUserID, err := GetUserIDFromAuthCookieOrSetNew(w, request)

		require.NoError(t, err)
		assert.Equal(t, userID, gotUserID)
	})

	t.Run("Set new cookie", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		gotUserID, err := GetUserIDFromAuthCookieOrSetNew(w, request)

		require.NoError(t, err)
		assert.NotEqual(t, gotUserID, uuid.Nil)
	})
}
