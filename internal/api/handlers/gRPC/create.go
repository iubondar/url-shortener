package grpc

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/auth"
	"github.com/iubondar/url-shortener/proto"
)

// TODO использовать существующий интерфейс
// URLSaver определяет интерфейс для сохранения URL в хранилище.
type URLSaver interface {
	// SaveURL сохраняет URL в хранилище.
	// Возвращает короткий идентификатор, флаг существования и ошибку.
	SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error)
}

type CreateIDHandler struct {
	proto.UnimplementedCreateIDHandlerServer

	saver   URLSaver // репозиторий для хранения URL
	baseURL string   // базовый URL для формирования сокращенных ссылок
}

func NewCreateIDGRPCHandler(saver URLSaver, baseURL string) *CreateIDHandler {
	return &CreateIDHandler{saver: saver, baseURL: baseURL}
}

func (h *CreateIDHandler) CreateID(ctx context.Context, req *proto.CreateIDRequest) (*proto.CreateIDResponse, error) {
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
	id, exists, err := h.saver.SaveURL(ctx, userID, parsedURL.String())
	if err != nil {
		response.Error = "Can't save URL"
		return &response, nil
	}

	if exists {
		response.Error = "URL already exists"
	} else {
		// Формируем сокращенный URL
		baseURL := strings.TrimSuffix(strings.TrimPrefix(h.baseURL, "http://"), "/")
		response.Result = fmt.Sprintf("http://%s/%s", baseURL, id)
	}

	return &response, nil
}
