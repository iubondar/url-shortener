package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/iubondar/url-shortener/internal/app/strings"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

	var records []URLRecord = []URLRecord{}
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

func (frepo *FileRepository) SaveURL(url string) (id string, exists bool, err error) {
	// Если URL уже был сохранён - возвращаем имеющееся значение
	for _, rec := range frepo.records {
		if rec.OriginalURL == url {
			return rec.ShortURL, true, nil
		}
	}

	// создаём идентификатор и добавляем запись
	id = strings.RandString(idLength)
	uuid := strconv.Itoa(frepo.nextID())
	record := URLRecord{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: url,
	}
	frepo.records = append(frepo.records, record)

	// сохраняем изменения на диск
	frepo.saveToFile(&record)

	return id, false, nil
}

func (frepo FileRepository) RetrieveURL(id string) (url string, err error) {
	for _, rec := range frepo.records {
		if rec.ShortURL == id {
			return rec.OriginalURL, nil
		}
	}

	return "", ErrorNotFound
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

func (frepo FileRepository) saveToFile(record *URLRecord) error {
	file, err := os.OpenFile(frepo.fPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	return json.NewEncoder(file).Encode(&record)
}
