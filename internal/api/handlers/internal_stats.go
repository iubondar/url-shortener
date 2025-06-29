package handlers

import (
	"context"
	"encoding/json"
	"net"
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
	trustedSubnet  string         // подсеть, с которой разрешен доступ к статистике
}

// NewInternalStatsHandler создает новый экземпляр InternalStatsHandler.
// Принимает репозиторий для получения статистики.
func NewInternalStatsHandler(statsRetriever StatsRetriever, trustedSubnet string) InternalStatsHandler {
	return InternalStatsHandler{statsRetriever: statsRetriever, trustedSubnet: trustedSubnet}
}

func (h InternalStatsHandler) GetStats(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	if h.trustedSubnet == "" {
		http.Error(res, "Trusted subnet is not set - access denied", http.StatusForbidden)
		return
	}

	ipStr := req.Header.Get("X-Real-IP")
	if ipStr == "" {
		http.Error(res, "X-Real-IP header is required", http.StatusForbidden)
		return
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		http.Error(res, "Invalid X-Real-IP header", http.StatusForbidden)
		return
	}

	// trustedSubnet - это подсеть в CIDR-нотации
	_, subnet, err := net.ParseCIDR(h.trustedSubnet)
	if err != nil {
		http.Error(res, "Invalid trusted subnet", http.StatusInternalServerError)
		return
	}

	if !subnet.Contains(ip) {
		http.Error(res, "Access denied - IP not in trusted subnet", http.StatusForbidden)
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
