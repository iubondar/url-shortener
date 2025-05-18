package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фиксированный UUID для тестов
var testUUID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

// ExampleGetUserIDFromAuthCookieOrSetNew демонстрирует пример получения идентификатора пользователя из cookie.
// Пример показывает, как получить существующий идентификатор пользователя.
func ExampleGetUserIDFromAuthCookieOrSetNew() {
	// Создаем тестовый HTTP запрос с существующей cookie
	request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	authCookie, _ := NewAuthCookie(testUUID)
	request.AddCookie(authCookie)
	w := httptest.NewRecorder()

	// Получаем идентификатор пользователя
	gotUserID, _ := GetUserIDFromAuthCookieOrSetNew(w, request)

	// Выводим идентификатор пользователя
	fmt.Println(gotUserID)
	// Output: 123e4567-e89b-12d3-a456-426614174000
}

// ExampleNewAuthCookie демонстрирует пример создания новой cookie с токеном аутентификации.
// Пример показывает, как создать cookie для нового пользователя.
func ExampleNewAuthCookie() {
	// Создаем cookie с фиксированным идентификатором пользователя
	cookie, _ := NewAuthCookie(testUUID)

	// Выводим имя cookie
	fmt.Println(cookie.Name)
	// Output: Authorization
}

// ExampleGetUserID демонстрирует пример извлечения идентификатора пользователя из токена.
// Пример показывает, как получить идентификатор пользователя из JWT токена.
func ExampleGetUserID() {
	// Создаем токен с фиксированным идентификатором пользователя
	tokenString, _ := buildJWTString(testUUID)

	// Получаем идентификатор пользователя из токена
	gotUserID, _ := GetUserID(tokenString)

	// Выводим идентификатор пользователя
	fmt.Println(gotUserID)
	// Output: 123e4567-e89b-12d3-a456-426614174000
}

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
