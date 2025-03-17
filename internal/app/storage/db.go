package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	SQLDB *sql.DB
	Repo  Repository
}

func NewDB(dsn string) (db *DB, err error) {
	pgx, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	repo, err := NewPGRepository(context.Background(), pgx)
	if err != nil {
		return nil, err
	}

	return &DB{
		SQLDB: pgx,
		Repo:  repo,
	}, nil
}
