// Package storage предоставляет интерфейсы и реализации для хранения и управления URL-ссылками.
package pg

import (
	"database/sql"

	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// DB представляет соединение с базой данных
type DB struct {
	SQLDB *sql.DB // соединение с базой данных
}

// NewDB создает новое соединение с базой данных и инициализирует репозиторий.
// Выполняет миграции базы данных при первом запуске.
// Принимает строку подключения к базе данных.
// Возвращает указатель на DB и ошибку, если она возникла.
func NewDB(dsn string) (db *DB, err error) {
	pgx, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.Up(pgx, "migrations"); err != nil {
		return nil, err
	}

	return &DB{
		SQLDB: pgx,
	}, nil
}
