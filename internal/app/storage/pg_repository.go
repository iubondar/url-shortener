package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/iubondar/url-shortener/internal/app/storage/queries"
	"github.com/iubondar/url-shortener/internal/app/strings"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type PGRepository struct {
	db *sql.DB
}

func NewPGRepository(db *sql.DB) (*PGRepository, error) {
	return &PGRepository{
		db: db,
	}, nil
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

func (repo *PGRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	row := repo.db.QueryRowContext(ctx, queries.GetOriginalURL, id)

	err = row.Scan(&url)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrorNotFound
	}

	return url, err
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

func (repo *PGRepository) RetrieveUserURLs(ctx context.Context, userID uuid.UUID) (URLPairs []URLPair, err error) {
	rows, err := repo.db.QueryContext(ctx, queries.GetUserUrls, userID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return []URLPair{}, nil
	} else if err != nil {
		return nil, err
	}

	defer rows.Close()

	URLPairs = []URLPair{}
	for rows.Next() {
		var pair URLPair
		err = rows.Scan(&pair.ShortURL, &pair.OriginalURL)
		if err != nil {
			return nil, err
		}
		URLPairs = append(URLPairs, pair)
	}
	return URLPairs, nil
}
