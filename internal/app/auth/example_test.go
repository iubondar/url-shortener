package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"
)

// Фиксированный UUID для тестов
var testUUID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

// Example демонстрирует пример получения идентификатора пользователя из cookie.
// Пример показывает, как получить существующий идентификатор пользователя.
func Example() {
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

// Example_cookie демонстрирует пример создания новой cookie с токеном аутентификации.
// Пример показывает, как создать cookie для нового пользователя.
func Example_cookie() {
	// Создаем cookie с фиксированным идентификатором пользователя
	cookie, _ := NewAuthCookie(testUUID)

	// Выводим имя cookie
	fmt.Println(cookie.Name)
	// Output: Authorization
}

// Example_token демонстрирует пример извлечения идентификатора пользователя из токена.
// Пример показывает, как получить идентификатор пользователя из JWT токена.
func Example_token() {
	// Создаем токен с фиксированным идентификатором пользователя
	tokenString, _ := buildJWTString(testUUID)

	// Получаем идентификатор пользователя из токена
	gotUserID, _ := GetUserID(tokenString)

	// Выводим идентификатор пользователя
	fmt.Println(gotUserID)
	// Output: 123e4567-e89b-12d3-a456-426614174000
}
