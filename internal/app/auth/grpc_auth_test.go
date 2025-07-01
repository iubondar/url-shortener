package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestGRPCAuthInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectNewToken bool
		expectError    bool
	}{
		{
			name:           "No token provided",
			token:          "",
			expectNewToken: true,
			expectError:    false,
		},
		{
			name:           "Valid token provided",
			token:          "",
			expectNewToken: false,
			expectError:    false,
		},
		{
			name:           "Invalid token provided",
			token:          "invalid_token",
			expectNewToken: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем interceptor
			interceptor := GRPCAuthInterceptor()

			// Создаем контекст с metadata
			var ctx context.Context
			if tt.token != "" {
				md := metadata.New(map[string]string{
					AuthMetadataKey: tt.token,
				})
				ctx = metadata.NewIncomingContext(context.Background(), md)
			} else {
				ctx = context.Background()
			}

			// Создаем mock handler
			handlerCalled := false
			var handlerUserID uuid.UUID
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				userID, err := GetUserIDFromContext(ctx)
				if err == nil {
					handlerUserID = userID
				}
				return "response", nil
			}

			// Вызываем interceptor
			resp, err := interceptor(ctx, "request", nil, handler)

			// Проверяем результаты
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, handlerCalled)
			assert.Equal(t, "response", resp)
			assert.NotEqual(t, uuid.Nil, handlerUserID)
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		expectError  bool
	}{
		{
			name: "UserID in context",
			setupContext: func() context.Context {
				userID := uuid.New()
				return context.WithValue(context.Background(), UserIDContextKey{}, userID)
			},
			expectError: false,
		},
		{
			name: "No UserID in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			userID, err := GetUserIDFromContext(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, userID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			}
		})
	}
}

func TestGRPCAuthInterceptorWithValidToken(t *testing.T) {
	// Создаем валидный токен
	userID := uuid.New()
	token, err := buildJWTString(userID)
	require.NoError(t, err)

	// Создаем interceptor
	interceptor := GRPCAuthInterceptor()

	// Создаем контекст с валидным токеном
	md := metadata.New(map[string]string{
		AuthMetadataKey: token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Создаем mock handler
	handlerCalled := false
	var handlerUserID uuid.UUID
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		userID, err := GetUserIDFromContext(ctx)
		if err == nil {
			handlerUserID = userID
		}
		return "response", nil
	}

	// Вызываем interceptor
	resp, err := interceptor(ctx, "request", nil, handler)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "response", resp)
	assert.Equal(t, userID, handlerUserID) // Должен быть тот же userID, что и в токене
}
