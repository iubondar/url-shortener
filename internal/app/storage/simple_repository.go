package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/strings"
)

const idLength int = 8

type Record struct {
	ShortURL    string
	OriginalURL string
	UserID      uuid.UUID
}
type SimpleRepository struct {
	Records []Record
}

func NewSimpleRepository() SimpleRepository {
	return SimpleRepository{
		Records: []Record{},
	}
}

func (repo SimpleRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	id, err = repo.RetrieveID(url)
	if err == nil && len(id) > 0 {
		return id, true, nil
	}

	// создаём идентификатор и сохраняем URL
	id = strings.RandString(idLength)
	repo.Records = append(
		repo.Records,
		Record{
			ShortURL:    id,
			OriginalURL: url,
			UserID:      userID,
		},
	)

	return id, false, nil
}

func (repo SimpleRepository) CheckStatus(ctx context.Context) error {
	// Статус всегда ок
	return nil
}

func (repo SimpleRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
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

func (repo SimpleRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	for _, r := range repo.Records {
		if r.ShortURL == id {
			return r.OriginalURL, nil
		}
	}

	return "", ErrorNotFound
}

func (repo SimpleRepository) RetrieveID(url string) (id string, err error) {
	for _, r := range repo.Records {
		if r.OriginalURL == id {
			return r.ShortURL, nil
		}
	}

	return "", ErrorNotFound
}
