package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/iubondar/url-shortener/internal/app/storage"
)

// RetrieveURLHandler обрабатывает запросы на получение оригинального URL по сокращенному идентификатору.
// Выполняет перенаправление на оригинальный URL или возвращает ошибку, если URL не найден или удален.
type RetrieveURLHandler struct {
	repo storage.Repository // репозиторий для хранения URL
}

// NewRetrieveURLHandler создает новый экземпляр RetrieveURLHandler.
// Принимает репозиторий для хранения URL.
func NewRetrieveURLHandler(repo storage.Repository) RetrieveURLHandler {
	return RetrieveURLHandler{
		repo: repo,
	}
}

// RetrieveURL обрабатывает HTTP GET запрос для получения оригинального URL.
// Принимает сокращенный идентификатор в параметре пути.
// Возвращает:
// - 307 Temporary Redirect с оригинальным URL в заголовке Location при успехе
// - 410 Gone если URL был удален
// - 400 Bad Request если URL не найден или параметр id отсутствует
func (handler RetrieveURLHandler) RetrieveURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(req, "id")
	if len(id) == 0 {
		http.Error(res, "Can't find id parameter in query path", http.StatusBadRequest)
		return
	}

	record, err := handler.repo.RetrieveByShortURL(req.Context(), id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if record.IsDeleted {
		res.WriteHeader(http.StatusGone)
	} else {
		res.Header().Add("Location", record.OriginalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
