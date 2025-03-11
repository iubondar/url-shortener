package storage

import (
	"context"

	"github.com/iubondar/url-shortener/internal/app/strings"
)

const idLength int = 8

type SimpleRepository struct {
	UrlsToIds map[string]string
	IdsToURLs map[string]string
}

func NewSimpleRepository() SimpleRepository {
	return SimpleRepository{
		UrlsToIds: make(map[string]string),
		IdsToURLs: make(map[string]string),
	}
}

func (repo SimpleRepository) SaveURL(ctx context.Context, url string) (id string, exists bool, err error) {
	id, ok := repo.UrlsToIds[url]
	if ok {
		// URL уже был сохранён - возвращаем имеющееся значение
		return id, true, nil
	}

	// создаём идентификатор и сохраняем URL
	id = strings.RandString(idLength)
	repo.UrlsToIds[url] = id
	repo.IdsToURLs[id] = url

	return id, false, nil
}

func (repo SimpleRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	url, ok := repo.IdsToURLs[id]
	if !ok {
		return "", ErrorNotFound
	}

	return url, nil
}

func (repo SimpleRepository) CheckStatus(ctx context.Context) error {
	// Статус всегда ок
	return nil
}

func (repo SimpleRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	ids = make([]string, 0)
	for _, url := range urls {
		id, _, err := repo.SaveURL(ctx, url)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (repo SimpleRepository) RetrieveID(url string) (id string, err error) {
	id, ok := repo.UrlsToIds[url]
	if !ok {
		return "", ErrorNotFound
	}

	return id, nil
}
