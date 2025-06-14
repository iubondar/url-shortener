// Package storage предоставляет интерфейсы и реализации для хранения и управления URL-ссылками.
package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/models"
	"github.com/iubondar/url-shortener/internal/app/storage/queries"
	"github.com/iubondar/url-shortener/internal/app/strings"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// deleteIn представляет структуру для удаления URL.
type deleteIn struct {
	shortURL string    // короткий идентификатор URL
	userID   uuid.UUID // идентификатор пользователя
}

const defaultDeletionInterval = 5 * time.Second

// PGRepository реализует хранилище URL на базе PostgreSQL.
// Поддерживает асинхронное удаление URL через очередь.
type PGRepository struct {
	db          *DB           // соединение с базой данных
	deleteQueue chan deleteIn // очередь для удаления URL
	insertStmt  *sql.Stmt     // подготовленный запрос для вставки URL
	getURLStmt  *sql.Stmt     // подготовленный запрос для получения URL
	deleteStmt  *sql.Stmt     // подготовленный запрос для удаления URL
}

// NewPGRepository создает новый экземпляр PGRepository.
// Принимает соединение с базой данных и интервал для асинхронного удаления. Если интервал не указан, используется значение по умолчанию.
// Возвращает указатель на PGRepository и ошибку, если она возникла.
func NewPGRepository(db *DB, deletionInterval time.Duration) (*PGRepository, error) {
	if deletionInterval == 0 {
		deletionInterval = defaultDeletionInterval
	}

	// Подготавливаем запросы
	insertStmt, err := db.SQLDB.Prepare(queries.InsertURL)
	if err != nil {
		return nil, fmt.Errorf("prepare insert statement: %w", err)
	}

	getURLStmt, err := db.SQLDB.Prepare(queries.GetShortURL)
	if err != nil {
		return nil, fmt.Errorf("prepare get URL statement: %w", err)
	}

	deleteStmt, err := db.SQLDB.Prepare(queries.DeleteUserURL)
	if err != nil {
		return nil, fmt.Errorf("prepare delete statement: %w", err)
	}

	instance := &PGRepository{
		db:          db,
		deleteQueue: make(chan deleteIn, 64),
		insertStmt:  insertStmt,
		getURLStmt:  getURLStmt,
		deleteStmt:  deleteStmt,
	}

	go instance.flushDeletions(deletionInterval)

	return instance, nil
}

// SaveURL сохраняет URL в базе данных.
// Если URL уже существует, возвращает его короткий идентификатор.
// Возвращает короткий идентификатор, флаг существования и ошибку.
func (repo *PGRepository) SaveURL(ctx context.Context, userID uuid.UUID, url string) (id string, exists bool, err error) {
	// создаём идентификатор и добавляем запись
	id = strings.RandString(8)
	_, err = repo.insertStmt.ExecContext(ctx, id, url, userID)
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

// getShortURLByOriginalURL получает короткий идентификатор по оригинальному URL.
// Возвращает короткий идентификатор и ошибку. Если URL не найден, возвращает пустую строку и nil.
func (repo *PGRepository) getShortURLByOriginalURL(ctx context.Context, url string) (shortURL string, err error) {
	err = repo.getURLStmt.QueryRowContext(ctx, url).Scan(&shortURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return shortURL, err
}

// RetrieveByShortURL получает запись по короткому идентификатору.
// Возвращает запись и ошибку. Если запись не найдена, возвращает ошибку ErrorNotFound.
func (repo *PGRepository) RetrieveByShortURL(ctx context.Context, shortURL string) (record models.Record, err error) {
	row := repo.db.SQLDB.QueryRowContext(ctx, queries.GetByShortURL, shortURL)

	err = row.Scan(&record.UserID, &record.ShortURL, &record.OriginalURL, &record.IsDeleted)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Record{}, models.ErrorNotFound
	}

	return
}

// CheckStatus проверяет состояние хранилища.
// Возвращает ошибку, если база данных недоступна.
func (repo *PGRepository) CheckStatus(ctx context.Context) error {
	return repo.db.SQLDB.PingContext(ctx)
}

// SaveURLs сохраняет массив URL в базе данных в одной транзакции.
// Если хотя бы один URL невалиден, откатывает транзакцию.
// Возвращает массив коротких идентификаторов и ошибку.
func (repo *PGRepository) SaveURLs(ctx context.Context, urls []string) (ids []string, err error) {
	tx, err := repo.db.SQLDB.Begin()
	if err != nil {
		return nil, err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				zap.L().Sugar().Errorf("error rolling back transaction: %v", rbErr)
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, queries.InsertURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			zap.L().Sugar().Errorf("error closing statement: %v", err)
		}
	}()

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
		id := strings.RandString(8)
		ids = append(ids, id)
		_, err = stmt.ExecContext(ctx, id, url, uuid.Nil)
		if err != nil {
			return nil, err
		}
	}

	return ids, tx.Commit()
}

// RetrieveUserURLs получает все URL пользователя.
// Возвращает массив записей и ошибку.
func (repo *PGRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (records []models.Record, err error) {
	rows, err := repo.db.SQLDB.QueryContext(ctx, queries.GetUserUrls, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Record{}, nil
		} else {
			return nil, err
		}
	}

	defer func() {
		if err := rows.Close(); err != nil {
			zap.L().Sugar().Errorf("error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var record models.Record
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

// DeleteByShortURLs помечает URL как удаленные.
// Добавляет URL в очередь для асинхронного удаления.
func (repo *PGRepository) DeleteByShortURLs(ctx context.Context, userID uuid.UUID, shortURLs []string) {
	for _, shortURL := range shortURLs {
		repo.deleteQueue <- deleteIn{shortURL: shortURL, userID: userID}
	}
}

// flushDeletions периодически сохраняет накопленные в очереди удаления в базе данных.
// Запускается в отдельной горутине при создании репозитория.
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

// markAsDeleted помечает URL как удаленные в базе данных.
// Выполняется в рамках транзакции.
func (repo *PGRepository) markAsDeleted(ctx context.Context, deletions ...deleteIn) error {
	tx, err := repo.db.SQLDB.Begin()
	if err != nil {
		return err
	}
	// если Commit будет раньше, то откат проигнорируется
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				zap.L().Sugar().Errorf("error rolling back transaction: %v", rbErr)
			}
		}
	}()

	stmt := tx.Stmt(repo.deleteStmt)
	defer func() {
		if err := stmt.Close(); err != nil {
			zap.L().Sugar().Errorf("error closing statement: %v", err)
		}
	}()

	for _, deleteIn := range deletions {
		_, err = stmt.ExecContext(ctx, deleteIn.userID, deleteIn.shortURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
