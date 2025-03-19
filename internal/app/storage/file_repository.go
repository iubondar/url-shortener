package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/strings"
)

type URLRecord struct {
	UUID        string    `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      uuid.UUID `json:"user_id"`
}

type FileRepository struct {
	fPath   string
	records []URLRecord
}

func NewFileRepository(fPath string) (*FileRepository, error) {
	// Создаём папки по указанному пути, если их ещё нет
	folderPath, _ := filepath.Split(fPath)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, os.ModePerm)
	}
	// Создаём файл, если его нет, или открываем на чтение
	file, err := os.OpenFile(fPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records = []URLRecord{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		err := json.Unmarshal(scanner.Bytes(), &record)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &FileRepository{
		fPath:   fPath,
		records: records,
	}, nil
}

func (frepo *FileRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	// Если URL уже был сохранён - возвращаем имеющееся значение
	record := frepo.getRecordByOriginalURL(url)
	if record != nil {
		return record.ShortURL, true, nil
	}

	// сохраняем изменения на диск
	record = frepo.addRecordForURL(url)

	frepo.appendToFile([]URLRecord{*record})

	return record.ShortURL, false, nil
}

// Вернём nil, если запись не найдена
func (frepo *FileRepository) getRecordByOriginalURL(originalURL string) *URLRecord {
	for _, rec := range frepo.records {
		if rec.OriginalURL == originalURL {
			return &rec
		}
	}

	return nil
}

func (frepo *FileRepository) addRecordForURL(url string) *URLRecord {
	// создаём идентификатор и добавляем запись
	id := strings.RandString(idLength)
	uuid := strconv.Itoa(frepo.nextID())
	record := URLRecord{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: url,
	}
	frepo.records = append(frepo.records, record)

	return &record
}

func (frepo FileRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	for _, rec := range frepo.records {
		if rec.ShortURL == id {
			return rec.OriginalURL, nil
		}
	}

	return "", ErrorNotFound
}

func (frepo FileRepository) CheckStatus(ctx context.Context) error {
	file, err := os.OpenFile(frepo.fPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func (frepo *FileRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	ids = make([]string, 0)
	newRecords := make([]URLRecord, 0)
	for _, url := range urls {
		record := frepo.getRecordByOriginalURL(url)
		if record != nil {
			ids = append(ids, record.ShortURL)
			continue
		}

		record = frepo.addRecordForURL(url)
		newRecords = append(newRecords, *record)
		ids = append(ids, record.ShortURL)
	}

	// сохраняем изменения на диск
	frepo.appendToFile(newRecords)

	return ids, nil
}

func (frepo FileRepository) nextID() int {
	if len(frepo.records) > 0 {
		last, err := strconv.Atoi(frepo.records[len(frepo.records)-1].UUID)
		if err != nil {
			log.Fatal(err)
		}
		return last + 1
	}
	return 1
}

func (frepo FileRepository) appendToFile(records []URLRecord) error {
	file, err := os.OpenFile(frepo.fPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	for _, record := range records {
		err = encoder.Encode(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (frepo FileRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (URLPairs []URLPair, err error) {
	URLPairs = make([]URLPair, 0)
	for _, r := range frepo.records {
		if r.UserID == userID {
			URLPairs = append(URLPairs, URLPair{ShortURL: r.ShortURL, OriginalURL: r.OriginalURL})
		}
	}
	return URLPairs, nil
}
