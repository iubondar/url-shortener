// Package storage предоставляет интерфейсы и реализации для хранения и управления URL-ссылками.
package simple

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage"
	"github.com/iubondar/url-shortener/internal/app/strings"
)

const idLength int = 8

// SimpleRepository реализует in-memory хранилище URL.
// Хранит все записи в памяти и не сохраняет их между запусками приложения.
type SimpleRepository struct {
	Records []storage.Record // массив записей URL
}

// NewSimpleRepository создает новый экземпляр SimpleRepository.
// Возвращает указатель на инициализированное хранилище.
func NewSimpleRepository() *SimpleRepository {
	return &SimpleRepository{
		Records: []storage.Record{},
	}
}

// SaveURL сохраняет URL в хранилище.
// Если URL уже существует, возвращает его короткий идентификатор.
// Возвращает короткий идентификатор, флаг существования и ошибку.
func (repo *SimpleRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	id, err = repo.RetrieveID(url)
	if err == nil && len(id) > 0 {
		return id, true, nil
	}

	// создаём идентификатор и сохраняем URL
	id = strings.RandString(idLength)
	repo.Records = append(
		repo.Records,
		storage.Record{
			ShortURL:    id,
			OriginalURL: url,
			UserID:      userID,
		},
	)

	return id, false, nil
}

// CheckStatus проверяет состояние хранилища.
// Для in-memory хранилища всегда возвращает nil.
func (repo SimpleRepository) CheckStatus(ctx context.Context) error {
	// Статус всегда ок
	return nil
}

// SaveURLs сохраняет массив URL в хранилище.
// Возвращает массив коротких идентификаторов и ошибку.
func (repo *SimpleRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	ids = make([]string, 0)
	for _, url := range urls {
		id, _, err := repo.SaveURL(ctx, uuid.Nil, url)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// RetrieveByShortURL получает запись по короткому идентификатору.
// Возвращает запись и ошибку.
func (repo SimpleRepository) RetrieveByShortURL(ctx context.Context, shortURL string) (record storage.Record, err error) {
	for _, r := range repo.Records {
		if r.ShortURL == shortURL {
			return r, nil
		}
	}

	return storage.Record{}, storage.ErrorNotFound
}

// RetrieveID получает короткий идентификатор по оригинальному URL.
// Возвращает короткий идентификатор и ошибку.
func (repo SimpleRepository) RetrieveID(url string) (id string, err error) {
	for _, r := range repo.Records {
		if r.OriginalURL == url {
			return r.ShortURL, nil
		}
	}

	return "", storage.ErrorNotFound
}

// RetrieveUserURLs получает все URL пользователя.
// Возвращает массив записей и ошибку.
func (repo SimpleRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []storage.Record, err error) {
	records = make([]storage.Record, 0)
	for _, r := range repo.Records {
		if r.UserID == userID {
			records = append(records, r)
		}
	}
	return records, nil
}

// DeleteByShortURLs помечает URL как удаленные.
// Принимает идентификатор пользователя и массив коротких идентификаторов.
func (repo *SimpleRepository) DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string) {
	for i, r := range repo.Records {
		if r.UserID == userID && slices.Contains(shortURLs, r.ShortURL) {
			r.IsDeleted = true
			repo.Records[i] = r
		}
	}
}
