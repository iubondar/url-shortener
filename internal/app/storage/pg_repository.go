package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage/queries"
	"github.com/iubondar/url-shortener/internal/app/strings"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type deleteIn struct {
	shortURL string
	userID   uuid.UUID
}

const defaultDeletionInterval = 5 * time.Second

type PGRepository struct {
	db          *sql.DB
	deleteQueue chan deleteIn
}

func NewPGRepository(db *sql.DB, deletionInterval time.Duration) (*PGRepository, error) {
	if deletionInterval == 0 {
		deletionInterval = defaultDeletionInterval
	}

	instance := &PGRepository{
		db:          db,
		deleteQueue: make(chan deleteIn, 1024),
	}

	go instance.flushDeletions(deletionInterval)

	return instance, nil
}

func (repo *PGRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	// создаём идентификатор и добавляем запись
	id = strings.RandString(idLength)
	_, err = repo.db.ExecContext(ctx, queries.InsertURL, id, url, userID)
	if err != nil {
		// Если URL уже был сохранён - возвращаем имеющееся значение
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {

			shortURL, err := repo.getShortURLByOriginalURL(ctx, url)
			if err != nil {
				zap.L().Sugar().Debugln("Error getting short URL:", err.Error())
				return "", false, err
			}

			if len(shortURL) > 0 {
				return shortURL, true, nil
			}
		}

		// Другая ошибка
		zap.L().Sugar().Debugln("Error insert new URL:", err.Error())
		return "", false, err
	}

	return id, false, nil
}

// Возвращает короткий URL если он уже есть в БД, иначе пустую строку
func (repo *PGRepository) getShortURLByOriginalURL(ctx context.Context, url string) (shortURL string, err error) {
	row := repo.db.QueryRowContext(ctx, queries.GetShortURL, url)

	err = row.Scan(&shortURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return shortURL, err
}

func (repo *PGRepository) RetrieveByShortURL(ctx context.Context, shortURL string) (record Record, err error) {
	row := repo.db.QueryRowContext(ctx, queries.GetByShortURL, shortURL)

	err = row.Scan(&record.UserID, &record.ShortURL, &record.OriginalURL, &record.IsDeleted)

	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, ErrorNotFound
	}

	return
}

func (repo *PGRepository) CheckStatus(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

// Сохраняем массив URL в одной транзакции
// Если хотя бы один из URL не валиден - откатываем транзакцию и возвращаем ошибку
func (repo *PGRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, queries.InsertURL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ids = make([]string, 0)
	for _, url := range urls {
		// Ищем в БД сохранённый URL
		existedURL, err := repo.getShortURLByOriginalURL(ctx, url)
		if err != nil {
			return nil, err
		}
		if len(existedURL) > 0 {
			ids = append(ids, existedURL)
			continue
		}

		// Сохраняем URL
		id := strings.RandString(idLength)
		ids = append(ids, id)
		_, err = stmt.ExecContext(ctx, id, url, uuid.Nil)
		if err != nil {
			return nil, err
		}
	}

	return ids, tx.Commit()
}

func (repo *PGRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []Record, err error) {
	rows, err := repo.db.QueryContext(ctx, queries.GetUserUrls, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Record{}, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		var record Record
		err = rows.Scan(&record.UserID, &record.ShortURL, &record.OriginalURL, &record.IsDeleted)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %s", err.Error())
	}

	return records, nil
}

func (repo *PGRepository) DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string) {
	for _, shortURL := range shortURLs {
		repo.deleteQueue <- deleteIn{shortURL: shortURL, userID: userID}
	}
}

// flushDeletions периодически сохраняет накопленные в очереди удаления в БД
func (repo *PGRepository) flushDeletions(deletionInterval time.Duration) {
	ticker := time.NewTicker(deletionInterval)

	var deletions []deleteIn

	for {
		select {
		case deleteIn := <-repo.deleteQueue:
			// добавим сообщение в слайс для последующего удаления
			deletions = append(deletions, deleteIn)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(deletions) == 0 {
				continue
			}
			// сохраним все пришедшие сообщения одновременно
			err := repo.markAsDeleted(context.Background(), deletions...)
			if err != nil {
				zap.L().Sugar().Debugln("cannot mark deletions:", err.Error())
			}
			// сотрём успешно отосланные сообщения
			deletions = nil
		}
	}
}

func (repo *PGRepository) markAsDeleted(ctx context.Context, deletions ...deleteIn) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, queries.DeleteUserURL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, deleteIn := range deletions {
		_, err = stmt.ExecContext(ctx, deleteIn.userID, deleteIn.shortURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
