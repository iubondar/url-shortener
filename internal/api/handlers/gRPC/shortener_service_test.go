package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/api/handlers/gRPC/mocks"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/iubondar/url-shortener/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/peer"
)

const testURL = "https://practicum.yandex.ru"
const testBaseURL = "127.0.0.1:8080"
const testTrustedSubnet = "192.168.1.0/24"

func TestShortenerService_CreateID(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		url            string
		setupMock      func(*mocks.MockRepository)
		expectedResult string
		expectedError  string
	}{
		{
			name: "Positive test",
			url:  testURL,
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().SaveURL(gomock.Any(), userID, testURL).Return("abc123", false, nil)
			},
			expectedResult: "http://127.0.0.1:8080/abc123",
			expectedError:  "",
		},
		{
			name: "URL already exists",
			url:  testURL,
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().SaveURL(gomock.Any(), userID, testURL).Return("abc123", true, nil)
			},
			expectedResult: "",
			expectedError:  "URL already exists",
		},
		{
			name: "Invalid URL",
			url:  "invalid-url",
			setupMock: func(m *mocks.MockRepository) {
				// Mock не вызывается для невалидного URL
			},
			expectedResult: "",
			expectedError:  "URL is not valid",
		},
		{
			name: "Repository error",
			url:  testURL,
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().SaveURL(gomock.Any(), userID, testURL).Return("", false, errors.New("db error"))
			},
			expectedResult: "",
			expectedError:  "Can't save URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

			// Создаем контекст с userID
			ctx := context.WithValue(context.Background(), auth.UserIDContextKey{}, userID)

			req := &proto.CreateIDRequest{Url: tt.url}
			resp, err := service.CreateID(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, resp.Result)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}

func TestShortenerService_CreateID_AuthenticationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Контекст без userID
	ctx := context.Background()

	req := &proto.CreateIDRequest{Url: testURL}
	resp, err := service.CreateID(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "", resp.Result)
	assert.Contains(t, resp.Error, "Authentication error")
}

func TestShortenerService_ShortenBatch(t *testing.T) {
	tests := []struct {
		name          string
		items         []*proto.BatchItem
		setupMock     func(*mocks.MockRepository)
		expectedItems int
		expectedError string
	}{
		{
			name: "Positive test",
			items: []*proto.BatchItem{
				{CorrelationId: "1", OriginalUrl: testURL},
				{CorrelationId: "2", OriginalUrl: "https://example.com"},
			},
			setupMock: func(m *mocks.MockRepository) {
				urls := []string{testURL, "https://example.com"}
				ids := []string{"abc123", "def456"}
				m.EXPECT().SaveURLs(gomock.Any(), urls).Return(ids, nil)
			},
			expectedItems: 2,
			expectedError: "",
		},
		{
			name: "Invalid URL in batch",
			items: []*proto.BatchItem{
				{CorrelationId: "1", OriginalUrl: "invalid-url"},
			},
			setupMock: func(m *mocks.MockRepository) {
				// Mock не вызывается для невалидного URL
			},
			expectedItems: 0,
			expectedError: "URL is not valid: parse \"invalid-url\": invalid URI for request",
		},
		{
			name: "Repository error",
			items: []*proto.BatchItem{
				{CorrelationId: "1", OriginalUrl: testURL},
			},
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().SaveURLs(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedItems: 0,
			expectedError: "Can't save URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

			req := &proto.ShortenBatchRequest{Items: tt.items}
			resp, err := service.ShortenBatch(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedItems, len(resp.Items))
			assert.Equal(t, tt.expectedError, resp.Error)

			if tt.expectedItems > 0 {
				assert.Contains(t, resp.Items[0].ShortUrl, "http://127.0.0.1:8080/")
			}
		})
	}
}

func TestShortenerService_GetUserURLs(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func(*mocks.MockRepository)
		expectedItems int
		expectedError string
	}{
		{
			name: "Positive test",
			setupMock: func(m *mocks.MockRepository) {
				records := []models.Record{
					{ShortURL: "abc123", OriginalURL: testURL, UserID: userID},
					{ShortURL: "def456", OriginalURL: "https://example.com", UserID: userID},
				}
				m.EXPECT().RetrieveUserURLs(gomock.Any(), userID).Return(records, nil)
			},
			expectedItems: 2,
			expectedError: "",
		},
		{
			name: "Empty user URLs",
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().RetrieveUserURLs(gomock.Any(), userID).Return([]models.Record{}, nil)
			},
			expectedItems: 0,
			expectedError: "",
		},
		{
			name: "Repository error",
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().RetrieveUserURLs(gomock.Any(), userID).Return(nil, errors.New("db error"))
			},
			expectedItems: 0,
			expectedError: "Can't retrieve user URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

			// Создаем контекст с userID
			ctx := context.WithValue(context.Background(), auth.UserIDContextKey{}, userID)

			req := &proto.GetUserURLsRequest{}
			resp, err := service.GetUserURLs(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedItems, len(resp.Items))
			assert.Equal(t, tt.expectedError, resp.Error)

			if tt.expectedItems > 0 {
				assert.Contains(t, resp.Items[0].ShortUrl, "http://127.0.0.1:8080/")
				assert.Equal(t, testURL, resp.Items[0].OriginalUrl)
			}
		})
	}
}

func TestShortenerService_GetUserURLs_AuthenticationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Контекст без userID
	ctx := context.Background()

	req := &proto.GetUserURLsRequest{}
	resp, err := service.GetUserURLs(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, 0, len(resp.Items))
	assert.Contains(t, resp.Error, "Authentication error")
}

func TestShortenerService_DeleteURLs(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		urls      []string
		setupMock func(*mocks.MockRepository)
	}{
		{
			name: "Positive test",
			urls: []string{"abc123", "def456"},
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().DeleteByShortURLs(gomock.Any(), userID, []string{"abc123", "def456"})
			},
		},
		{
			name: "Empty URLs list",
			urls: []string{},
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().DeleteByShortURLs(gomock.Any(), userID, []string{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

			// Создаем контекст с userID
			ctx := context.WithValue(context.Background(), auth.UserIDContextKey{}, userID)

			req := &proto.DeleteURLsRequest{Urls: tt.urls}
			resp, err := service.DeleteURLs(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, "", resp.Error)
		})
	}
}

func TestShortenerService_DeleteURLs_AuthenticationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Контекст без userID
	ctx := context.Background()

	req := &proto.DeleteURLsRequest{Urls: []string{"abc123"}}
	resp, err := service.DeleteURLs(ctx, req)

	require.NoError(t, err)
	assert.Contains(t, resp.Error, "Authentication error")
}

func TestShortenerService_Ping(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockRepository)
		expectedStatus proto.Status
		expectedError  string
	}{
		{
			name: "Positive test",
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().CheckStatus(gomock.Any()).Return(nil)
			},
			expectedStatus: proto.Status_STATUS_OK,
			expectedError:  "",
		},
		{
			name: "Repository error",
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().CheckStatus(gomock.Any()).Return(errors.New("db connection failed"))
			},
			expectedStatus: proto.Status_STATUS_ERROR,
			expectedError:  "db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

			req := &proto.PingRequest{}
			resp, err := service.Ping(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}

func TestShortenerService_GetStats(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		clientIP      string
		setupMock     func(*mocks.MockRepository)
		expectedUrls  int32
		expectedUsers int32
		expectedError string
	}{
		{
			name:          "Positive test",
			trustedSubnet: testTrustedSubnet,
			clientIP:      "192.168.1.100",
			setupMock: func(m *mocks.MockRepository) {
				stats := models.Stats{URLsCount: 100, UsersCount: 50}
				m.EXPECT().GetStats(gomock.Any()).Return(stats, nil)
			},
			expectedUrls:  100,
			expectedUsers: 50,
			expectedError: "",
		},
		{
			name:          "Trusted subnet not set",
			trustedSubnet: "",
			clientIP:      "192.168.1.100",
			setupMock: func(m *mocks.MockRepository) {
				// Mock не вызывается
			},
			expectedUrls:  0,
			expectedUsers: 0,
			expectedError: "Trusted subnet is not set - access denied",
		},
		{
			name:          "IP not in trusted subnet",
			trustedSubnet: testTrustedSubnet,
			clientIP:      "10.0.0.1",
			setupMock: func(m *mocks.MockRepository) {
				// Mock не вызывается
			},
			expectedUrls:  0,
			expectedUsers: 0,
			expectedError: "Access denied - IP not in trusted subnet",
		},
		{
			name:          "Repository error",
			trustedSubnet: testTrustedSubnet,
			clientIP:      "192.168.1.100",
			setupMock: func(m *mocks.MockRepository) {
				m.EXPECT().GetStats(gomock.Any()).Return(models.Stats{}, errors.New("db error"))
			},
			expectedUrls:  0,
			expectedUsers: 0,
			expectedError: "Can't get stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewShortenerService(mockRepo, testBaseURL, tt.trustedSubnet)

			// Создаем контекст с peer информацией
			ctx := context.Background()
			if tt.clientIP != "" {
				addr := &net.TCPAddr{IP: net.ParseIP(tt.clientIP), Port: 12345}
				ctx = peer.NewContext(ctx, &peer.Peer{Addr: addr})
			}

			req := &proto.GetStatsRequest{}
			resp, err := service.GetStats(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedUrls, resp.Urls)
			assert.Equal(t, tt.expectedUsers, resp.Users)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}

func TestShortenerService_GetStats_NoPeerInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Контекст без peer информации
	ctx := context.Background()

	req := &proto.GetStatsRequest{}
	resp, err := service.GetStats(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Urls)
	assert.Equal(t, int32(0), resp.Users)
	assert.Equal(t, "Unable to get peer information", resp.Error)
}

func TestShortenerService_GetStats_InvalidPeerAddr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Контекст с невалидным peer адресом
	addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345}
	ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: addr})

	req := &proto.GetStatsRequest{}
	resp, err := service.GetStats(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Urls)
	assert.Equal(t, int32(0), resp.Users)
	assert.Equal(t, "Unable to get client IP address", resp.Error)
}

func TestShortenerService_GetStats_InvalidTrustedSubnet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := NewShortenerService(mockRepo, testBaseURL, "invalid-subnet")

	addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345}
	ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: addr})

	req := &proto.GetStatsRequest{}
	resp, err := service.GetStats(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Urls)
	assert.Equal(t, int32(0), resp.Users)
	assert.Equal(t, "Invalid trusted subnet", resp.Error)
}

// ExampleShortenerService_CreateID демонстрирует пример использования gRPC метода CreateID
func ExampleShortenerService_CreateID() {
	// Создаем мок репозитория
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	// Настраиваем ожидания
	userID := uuid.New()
	mockRepo.EXPECT().SaveURL(gomock.Any(), userID, testURL).Return("abc123", false, nil)

	// Создаем сервис
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Создаем контекст с userID
	ctx := context.WithValue(context.Background(), auth.UserIDContextKey{}, userID)

	// Создаем запрос
	req := &proto.CreateIDRequest{Url: testURL}

	// Вызываем метод
	resp, err := service.CreateID(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Выводим результат
	fmt.Printf("Short URL: %s\n", resp.Result)
	fmt.Printf("Error: %s\n", resp.Error)
	// Output:
	// Short URL: http://127.0.0.1:8080/abc123
	// Error:
}

// ExampleShortenerService_Ping демонстрирует пример использования gRPC метода Ping
func ExampleShortenerService_Ping() {
	// Создаем мок репозитория
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	// Настраиваем ожидания
	mockRepo.EXPECT().CheckStatus(gomock.Any()).Return(nil)

	// Создаем сервис
	service := NewShortenerService(mockRepo, testBaseURL, testTrustedSubnet)

	// Создаем запрос
	req := &proto.PingRequest{}

	// Вызываем метод
	resp, err := service.Ping(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Выводим результат
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Error: %s\n", resp.Error)
	// Output:
	// Status: STATUS_OK
	// Error:
}
