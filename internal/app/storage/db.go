package storage

import (
	"database/sql"

	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type DB struct {
	SQLDB *sql.DB
	Repo  Repository
}

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

	repo, err := NewPGRepository(pgx)
	if err != nil {
		return nil, err
	}

	return &DB{
		SQLDB: pgx,
		Repo:  repo,
	}, nil
}
