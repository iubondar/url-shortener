package storage

import "github.com/iubondar/url-shortener/internal/strings"

const idLength int = 8

type SimpleRepository struct {
	urlsToIds map[string]string
	idsToURLs map[string]string
}

func NewSimpleRepository() SimpleRepository {
	return SimpleRepository{
		urlsToIds: make(map[string]string),
		idsToURLs: make(map[string]string),
	}
}

func (rep SimpleRepository) SaveURL(url string) (id string, exists bool, err error) {
	id, ok := rep.urlsToIds[url]
	if ok {
		// URL уже был сохранён - возвращаем имеющееся значение
		return id, true, nil
	}

	// создаём идентификатор и сохраняем URL
	id = strings.RandString(idLength)
	rep.urlsToIds[url] = id
	rep.idsToURLs[id] = url

	return id, false, nil
}

func (rep SimpleRepository) RetrieveURL(id string) (url string, err error) {
	url, ok := rep.idsToURLs[id]
	if !ok {
		return "", ErrorURLNotFound
	}

	return url, nil
}
