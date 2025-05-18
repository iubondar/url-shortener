package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/strings"
)

// URLRecord представляет запись URL в файловом хранилище.
// Содержит основную информацию о URL и дополнительное поле UUID для внутренней идентификации.
type URLRecord struct {
	Record
	UUID string `json:"uuid"` // внутренний идентификатор записи
}

// FileRepository реализует файловое хранилище URL.
// Сохраняет все записи в JSON-файле и поддерживает их загрузку при инициализации.
type FileRepository struct {
	fPath   string      // путь к файлу хранилища
	records []URLRecord // массив записей URL
}

// NewFileRepository создает новый экземпляр FileRepository.
// Создает файл хранилища, если он не существует, и загружает существующие записи.
// Принимает путь к файлу хранилища.
// Возвращает указатель на FileRepository и ошибку, если она возникла.
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

// SaveURL сохраняет URL в файловом хранилище.
// Если URL уже существует, возвращает его короткий идентификатор.
// Возвращает короткий идентификатор, флаг существования и ошибку.
func (frepo *FileRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	// Если URL уже был сохранён - возвращаем имеющееся значение
	record := frepo.getRecordByOriginalURL(url)
	if record != nil {
		return record.ShortURL, true, nil
	}

	// сохраняем изменения на диск
	record = frepo.addRecordForURL(url, userID)

	frepo.appendToFile([]URLRecord{*record})

	return record.ShortURL, false, nil
}

// getRecordByOriginalURL ищет запись по оригинальному URL.
// Возвращает указатель на найденную запись или nil, если запись не найдена.
func (frepo *FileRepository) getRecordByOriginalURL(originalURL string) *URLRecord {
	for _, rec := range frepo.records {
		if rec.OriginalURL == originalURL {
			return &rec
		}
	}

	return nil
}

// addRecordForURL создает новую запись для URL и добавляет её в хранилище.
// Генерирует короткий идентификатор и внутренний UUID.
// Возвращает указатель на созданную запись.
func (frepo *FileRepository) addRecordForURL(url string, userID uuid.UUID) *URLRecord {
	// создаём идентификатор и добавляем запись
	id := strings.RandString(idLength)
	uuid := strconv.Itoa(frepo.nextID())
	record := URLRecord{
		UUID: uuid,
		Record: Record{
			ShortURL:    id,
			OriginalURL: url,
			UserID:      userID,
		},
	}
	frepo.records = append(frepo.records, record)

	return &record
}

// RetrieveByShortURL получает запись по короткому идентификатору.
// Возвращает запись и ошибку. Если запись не найдена, возвращает ошибку ErrorNotFound.
func (frepo FileRepository) RetrieveByShortURL(ctx context.Context, shortURL string) (record Record, err error) {
	for _, rec := range frepo.records {
		if rec.ShortURL == shortURL {
			return rec.Record, nil
		}
	}

	return Record{}, ErrorNotFound
}

// CheckStatus проверяет состояние файлового хранилища.
// Проверяет доступность файла для чтения.
// Возвращает ошибку, если файл недоступен.
func (frepo FileRepository) CheckStatus(ctx context.Context) error {
	file, err := os.OpenFile(frepo.fPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

// SaveURLs сохраняет массив URL в файловом хранилище.
// Возвращает массив коротких идентификаторов и ошибку.
func (frepo *FileRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	ids = make([]string, 0)
	newRecords := make([]URLRecord, 0)
	for _, url := range urls {
		record := frepo.getRecordByOriginalURL(url)
		if record != nil {
			ids = append(ids, record.ShortURL)
			continue
		}

		record = frepo.addRecordForURL(url, uuid.Nil)
		newRecords = append(newRecords, *record)
		ids = append(ids, record.ShortURL)
	}

	// сохраняем изменения на диск
	frepo.appendToFile(newRecords)

	return ids, nil
}

// nextID генерирует следующий внутренний идентификатор записи.
// Возвращает целочисленный идентификатор.
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

// appendToFile добавляет записи в конец файла хранилища.
// Записи сериализуются в JSON и записываются построчно.
// Возвращает ошибку, если запись в файл не удалась.
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

// RetrieveUserURLs получает все URL пользователя.
// Возвращает массив записей и ошибку.
func (frepo FileRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []Record, err error) {
	for _, r := range frepo.records {
		if r.UserID == userID {
			records = append(records, r.Record)
		}
	}
	return records, nil
}

// DeleteByShortURLs помечает URL как удаленные.
// Принимает идентификатор пользователя и массив коротких идентификаторов.
// Обновляет записи в памяти, но не сохраняет изменения на диск.
func (frepo FileRepository) DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string) {
	for i, r := range frepo.records {
		if r.UserID == userID && slices.Contains(shortURLs, r.ShortURL) {
			r.IsDeleted = true
			frepo.records[i] = r
		}
	}
}
