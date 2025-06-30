package grpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/iubondar/url-shortener/proto"
	"google.golang.org/grpc/peer"
)

// Repository определяет интерфейс для работы с хранилищем.
// Использует существующий интерфейс из HTTP хэндлеров.
type Repository interface {
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
	RetrieveByShortURL(ctx context.Context, shortURL string) (record models.Record, err error)
	RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []models.Record, err error)
	DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string)
	CheckStatus(ctx context.Context) error
	SaveURLs(ctx context.Context, urls []string) (ids []string, err error)
	GetStats(ctx context.Context) (stats models.Stats, err error)
}

// ShortenerService реализует gRPC сервис для сокращения URL.
// Объединяет все методы в одном сервисе.
type ShortenerService struct {
	proto.UnimplementedShortenerServer

	repo          Repository // репозиторий для работы с хранилищем
	baseURL       string     // базовый URL для формирования сокращенных ссылок
	trustedSubnet string     // подсеть, с которой разрешен доступ к статистике
}

// NewShortenerService создает новый экземпляр ShortenerService.
func NewShortenerService(repo Repository, baseURL string, trustedSubnet string) *ShortenerService {
	return &ShortenerService{
		repo:          repo,
		baseURL:       baseURL,
		trustedSubnet: trustedSubnet,
	}
}

// CreateID создает короткий ID для URL.
func (s *ShortenerService) CreateID(ctx context.Context, req *proto.CreateIDRequest) (*proto.CreateIDResponse, error) {
	var response proto.CreateIDResponse

	// Валидируем URL
	parsedURL, err := url.ParseRequestURI(req.Url)
	if err != nil {
		response.Error = "URL is not valid"
		return &response, nil
	}

	// Получаем userID из контекста (установлен interceptor'ом)
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		response.Error = "Authentication error: " + err.Error()
		return &response, nil
	}

	// Сохраняем URL
	id, exists, err := s.repo.SaveURL(ctx, userID, parsedURL.String())
	if err != nil {
		response.Error = "Can't save URL"
		return &response, nil
	}

	if exists {
		response.Error = "URL already exists"
	} else {
		// Формируем сокращенный URL
		baseURL := strings.TrimSuffix(strings.TrimPrefix(s.baseURL, "http://"), "/")
		response.Result = fmt.Sprintf("http://%s/%s", baseURL, id)
	}

	return &response, nil
}

// ShortenBatch выполняет пакетное сокращение URL.
func (s *ShortenerService) ShortenBatch(ctx context.Context, req *proto.ShortenBatchRequest) (*proto.ShortenBatchResponse, error) {
	var response proto.ShortenBatchResponse

	// Извлекаем и валидируем URL из запроса
	var urls []string
	for _, item := range req.Items {
		// Проверяем URL
		parsedURL, err := url.ParseRequestURI(item.OriginalUrl)
		if err != nil {
			response.Error = "URL is not valid: " + err.Error()
			return &response, nil
		}
		urls = append(urls, parsedURL.String())
	}

	// Сохраняем URL пакетно
	ids, err := s.repo.SaveURLs(ctx, urls)
	if err != nil {
		response.Error = "Can't save URLs"
		return &response, nil
	}

	// Формируем ответ
	baseURL := strings.TrimSuffix(strings.TrimPrefix(s.baseURL, "http://"), "/")
	for i, item := range req.Items {
		if i < len(ids) {
			response.Items = append(response.Items, &proto.BatchItemResponse{
				CorrelationId: item.CorrelationId,
				ShortUrl:      fmt.Sprintf("http://%s/%s", baseURL, ids[i]),
			})
		}
	}

	return &response, nil
}

// GetUserURLs возвращает список URL пользователя.
func (s *ShortenerService) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	var response proto.GetUserURLsResponse

	// Получаем userID из контекста
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		response.Error = "Authentication error: " + err.Error()
		return &response, nil
	}

	// Получаем URL пользователя
	records, err := s.repo.RetrieveUserURLs(ctx, userID)
	if err != nil {
		response.Error = "Can't retrieve user URLs"
		return &response, nil
	}

	// Формируем ответ
	baseURL := strings.TrimSuffix(strings.TrimPrefix(s.baseURL, "http://"), "/")
	for _, record := range records {
		response.Items = append(response.Items, &proto.UserURL{
			ShortUrl:    fmt.Sprintf("http://%s/%s", baseURL, record.ShortURL),
			OriginalUrl: record.OriginalURL,
		})
	}

	return &response, nil
}

// DeleteURLs удаляет URL пользователя.
func (s *ShortenerService) DeleteURLs(ctx context.Context, req *proto.DeleteURLsRequest) (*proto.DeleteURLsResponse, error) {
	var response proto.DeleteURLsResponse

	// Получаем userID из контекста
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		response.Error = "Authentication error: " + err.Error()
		return &response, nil
	}

	// Удаляем URL
	s.repo.DeleteByShortURLs(ctx, userID, req.Urls)

	return &response, nil
}

// Ping проверяет доступность сервиса.
func (s *ShortenerService) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	var response proto.PingResponse

	// Проверяем статус хранилища
	err := s.repo.CheckStatus(ctx)
	if err != nil {
		response.Status = proto.Status_STATUS_ERROR
		response.Error = err.Error()
		return &response, nil
	}

	response.Status = proto.Status_STATUS_OK
	return &response, nil
}

// GetStats возвращает статистику сервиса.
func (s *ShortenerService) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	var response proto.GetStatsResponse

	// Проверяем, установлена ли доверенная подсеть
	if s.trustedSubnet == "" {
		response.Error = "Trusted subnet is not set - access denied"
		return &response, nil
	}

	// Получаем информацию о клиенте из контекста
	p, ok := peer.FromContext(ctx)
	if !ok {
		response.Error = "Unable to get peer information"
		return &response, nil
	}

	// Извлекаем IP адрес из peer
	addr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		response.Error = "Unable to get client IP address"
		return &response, nil
	}

	ip := addr.IP
	if ip == nil {
		response.Error = "Invalid client IP address"
		return &response, nil
	}

	// Парсим доверенную подсеть в CIDR-нотации
	_, subnet, err := net.ParseCIDR(s.trustedSubnet)
	if err != nil {
		response.Error = "Invalid trusted subnet"
		return &response, nil
	}

	// Проверяем, находится ли IP в доверенной подсети
	if !subnet.Contains(ip) {
		response.Error = "Access denied - IP not in trusted subnet"
		return &response, nil
	}

	// Получаем статистику
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		response.Error = "Can't get stats"
		return &response, nil
	}

	response.Urls = int32(stats.URLsCount)
	response.Users = int32(stats.UsersCount)

	return &response, nil
}
