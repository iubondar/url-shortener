package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/models"
)

// StatsRetriever определяет интерфейс для получения статистики.
type StatsRetriever interface {
	GetStats(ctx context.Context) (stats models.Stats, err error)
}

// InternalStatsHandler обрабатывает запросы на получение внутренней статистики.
// Позволяет администратору получать статистику о количестве сокращенных URL и пользователей.
type InternalStatsHandler struct {
	statsRetriever StatsRetriever // репозиторий для получения статистики
}

// NewInternalStatsHandler создает новый экземпляр InternalStatsHandler.
// Принимает репозиторий для получения статистики.
func NewInternalStatsHandler(statsRetriever StatsRetriever) InternalStatsHandler {
	return InternalStatsHandler{statsRetriever: statsRetriever}
}

func (h InternalStatsHandler) GetStats(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.statsRetriever.GetStats(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(stats)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err := res.Write(resp); err != nil {
		http.Error(res, "Error writing response", http.StatusInternalServerError)
		return
	}
}
