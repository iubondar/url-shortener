package storage

import (
	"context"
	"database/sql"

	"github.com/iubondar/url-shortener/internal/app/strings"
	"go.uber.org/zap"
)

type PGRepository struct {
	db *sql.DB
}

func NewPGRepository(ctx context.Context, db *sql.DB) (*PGRepository, error) {
	err := createTableIfNeeded(ctx, db)

	if err != nil {
		return nil, err
	}

	return &PGRepository{
		db: db,
	}, nil
}

func createTableIfNeeded(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS urls ("+
			"id SERIAL PRIMARY KEY,"+
			"short_url VARCHAR(10) UNIQUE NOT NULL,"+
			"original_url VARCHAR(2048) UNIQUE NOT NULL"+
			");",
	)

	return err
}

func (repo *PGRepository) SaveURL(ctx context.Context, url string) (id string, exists bool, err error) {
	// Если URL уже был сохранён - возвращаем имеющееся значение
	shortURL, err := repo.getShortURLByOriginalURL(ctx, url)
	if err != nil {
		zap.L().Sugar().Debugln("Error getting short URL:", err.Error())
		return "", false, err
	}

	if len(shortURL) > 0 {
		return shortURL, true, nil
	}

	// создаём идентификатор и добавляем запись
	id = strings.RandString(idLength)
	err = repo.saveURL(ctx, id, url)
	if err != nil {
		zap.L().Sugar().Debugln("Error saving URL:", err.Error())
		return "", false, err
	}

	return id, false, nil
}

// Возвращает короткий URL если он уже есть в БД, иначе пустую строку
func (repo *PGRepository) getShortURLByOriginalURL(ctx context.Context, url string) (shortURL string, err error) {
	row := repo.db.QueryRowContext(ctx, "SELECT short_url from urls WHERE original_url = $1;", url)

	err = row.Scan(&shortURL)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return shortURL, err
}

func (repo *PGRepository) saveURL(ctx context.Context, shortURL string, originalURL string) error {
	_, err := repo.db.ExecContext(
		ctx,
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2);",
		shortURL,
		originalURL,
	)

	return err
}

func (repo *PGRepository) RetrieveURL(ctx context.Context, id string) (url string, err error) {
	row := repo.db.QueryRowContext(ctx, "SELECT original_url from urls WHERE short_url = $1;", id)

	err = row.Scan(&url)

	if err == sql.ErrNoRows {
		return "", ErrorNotFound
	}

	return url, err
}

func (repo *PGRepository) CheckStatus(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}
