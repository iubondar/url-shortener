// Package auth предоставляет функциональность для аутентификации пользователей.
// Использует JWT токены для хранения идентификатора пользователя в cookie.
package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthMetadataKey - ключ для передачи JWT токена в gRPC metadata
const AuthMetadataKey = "authorization"

// UserIDContextKey - ключ для передачи userID в context
type UserIDContextKey struct{}

// GRPCAuthInterceptor создает unary interceptor для аутентификации gRPC запросов.
// Извлекает JWT токен из metadata, валидирует его и добавляет userID в context.
// Если токен отсутствует или невалидный, создает новый токен и возвращает его в response metadata.
func GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Извлекаем metadata из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Получаем JWT токен из metadata
		authValues := md.Get(AuthMetadataKey)
		var tokenString string
		if len(authValues) > 0 {
			tokenString = authValues[0]
		}

		var userID uuid.UUID
		var err error

		// Если токен есть, пытаемся его валидировать
		if tokenString != "" {
			userID, err = GetUserID(tokenString)
			if err != nil {
				zap.L().Sugar().Debugln("Error getting user id from token, will create new. Message: ", err.Error())
				// Токен невалидный, создаем новый
				userID = uuid.New()
			}
		} else {
			zap.L().Sugar().Debugln("No auth token found, create new")
			// Токена нет, создаем новый
			userID = uuid.New()
		}

		// Добавляем userID в контекст
		ctxWithUserID := context.WithValue(ctx, UserIDContextKey{}, userID)

		// Вызываем следующий обработчик
		resp, err := handler(ctxWithUserID, req)
		if err != nil {
			return resp, err
		}

		// Если был создан новый токен, добавляем его в response metadata
		if tokenString == "" {
			newToken, err := buildJWTString(userID)
			if err != nil {
				zap.L().Sugar().Errorln("Error building new JWT token: ", err.Error())
				return resp, err
			}

			// Создаем response metadata с новым токеном
			responseMD := metadata.New(map[string]string{
				AuthMetadataKey: newToken,
			})

			// Отправляем metadata в ответе
			if err := grpc.SetHeader(ctx, responseMD); err != nil {
				zap.L().Sugar().Errorln("Error setting response metadata: ", err.Error())
			}
		}

		return resp, nil
	}
}

// GetUserIDFromContext извлекает идентификатор пользователя из контекста gRPC.
// Возвращает userID и ошибку, если userID не найден в контексте.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDContextKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("userID not found in context")
	}
	return userID, nil
}
