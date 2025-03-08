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

func (rep SimpleRepository) SaveURL(ctx context.Context, url string) (id string, exists bool, err error) {
	id, ok := rep.UrlsToIds[url]
	if ok {
		// URL уже был сохранён - возвращаем имеющееся значение
		return id, true, nil
	}

	// создаём идентификатор и сохраняем URL
	id = strings.RandString(idLength)
	rep.UrlsToIds[url] = id
	rep.IdsToURLs[id] = url

	return id, false, nil
}

func (rep SimpleRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	url, ok := rep.IdsToURLs[id]
	if !ok {
		return "", ErrorNotFound
	}

	return url, nil
}

func (rep SimpleRepository) CheckStatus(ctx context.Context) error {
	// Статус всегда ок
	return nil
}

func (rep SimpleRepository) RetrieveID(url string) (id string, err error) {
	id, ok := rep.UrlsToIds[url]
	if !ok {
		return "", ErrorNotFound
	}

	return id, nil
}
