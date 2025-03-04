package handlers

import (
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/storage"
)

type PingHandler struct {
	checker storage.StatusChecker
}

func NewPingHandler(checker storage.StatusChecker) PingHandler {
	return PingHandler{
		checker: checker,
	}
}

func (handler PingHandler) Ping(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	err := handler.checker.CheckStatus(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
