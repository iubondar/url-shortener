package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const secret_key = "supersecretkey"
const authCookieName = "Authorization"

// claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

func SetAuthCookie(res http.ResponseWriter, req *http.Request) (userID uuid.UUID, err error) {
	authCookie, err := req.Cookie(authCookieName)
	if err != nil {
		zap.L().Sugar().Debugln("No auth cookie found, set new")
		return setNewAuthCookie(res)
	}

	userID, err = getUserID(authCookie.Value)
	if err != nil {
		zap.L().Sugar().Debugln("Error getting user id from cookie, will set new. Message: ", err.Error())
		return setNewAuthCookie(res)
	}

	return userID, nil
}

func setNewAuthCookie(res http.ResponseWriter) (userID uuid.UUID, err error) {
	userID = uuid.New()

	jwtString, err := buildJWTString(userID)
	if err != nil {
		zap.L().Sugar().Debugln("Error building jwtString", err.Error())
		return uuid.Nil, err
	}

	cookie := &http.Cookie{
		Name:     authCookieName,
		Value:    jwtString,
		HttpOnly: true, // Prevents JavaScript access
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(res, cookie)

	return userID, nil
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func buildJWTString(userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secret_key))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func getUserID(tokenString string) (userID uuid.UUID, err error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret_key), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
