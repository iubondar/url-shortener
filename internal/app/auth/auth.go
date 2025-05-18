// Package auth предоставляет функциональность для аутентификации пользователей.
// Использует JWT токены для хранения идентификатора пользователя в cookie.
package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const secretKey = "supersecretkey"

// AuthCookieName - имя cookie для хранения токена аутентификации
const AuthCookieName = "Authorization"

// claims представляет структуру JWT токена.
// Включает стандартные поля JWT и идентификатор пользователя.
type claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID // идентификатор пользователя
}

// GetUserIDFromAuthCookieOrSetNew получает идентификатор пользователя из cookie или создает новый.
// Если cookie не существует или содержит невалидный токен, создает новый токен.
// Возвращает идентификатор пользователя и ошибку, если она возникла.
func GetUserIDFromAuthCookieOrSetNew(res http.ResponseWriter, req *http.Request) (userID uuid.UUID, err error) {
	authCookie, err := req.Cookie(AuthCookieName)
	if err != nil {
		zap.L().Sugar().Debugln("No auth cookie found, set new")
		return setNewAuthCookie(res)
	}

	userID, err = GetUserID(authCookie.Value)
	if err != nil {
		zap.L().Sugar().Debugln("Error getting user id from cookie, will set new. Message: ", err.Error())
		return setNewAuthCookie(res)
	}

	return userID, nil
}

// setNewAuthCookie создает новый токен аутентификации и устанавливает его в cookie.
// Возвращает новый идентификатор пользователя и ошибку, если она возникла.
func setNewAuthCookie(res http.ResponseWriter) (userID uuid.UUID, err error) {
	userID = uuid.New()

	authCookie, err := NewAuthCookie(userID)
	if err != nil {
		return uuid.Nil, err
	}

	http.SetCookie(res, authCookie)

	return userID, nil
}

// NewAuthCookie создает новую cookie с JWT токеном для указанного пользователя.
// Возвращает cookie и ошибку, если она возникла.
func NewAuthCookie(userID uuid.UUID) (authCookie *http.Cookie, err error) {
	jwtString, err := buildJWTString(userID)
	if err != nil {
		zap.L().Sugar().Debugln("Error building jwtString", err.Error())
		return nil, err
	}

	authCookie = &http.Cookie{
		Name:     AuthCookieName,
		Value:    jwtString,
		HttpOnly: true, // Prevents JavaScript access
		SameSite: http.SameSiteLaxMode,
	}

	return authCookie, nil
}

// buildJWTString создает JWT токен для указанного пользователя и возвращает его в виде строки.
// Использует алгоритм подписи HS256.
// Возвращает строку токена и ошибку, если она возникла.
func buildJWTString(userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// GetUserID извлекает идентификатор пользователя из JWT токена.
// Проверяет валидность токена и его подпись.
// Возвращает идентификатор пользователя и ошибку, если она возникла.
func GetUserID(tokenString string) (userID uuid.UUID, err error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
