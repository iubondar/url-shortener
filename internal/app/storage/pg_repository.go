package storage

import (
	"context"
	"database/sql"
)

type PGRepository struct {
	db *sql.DB
}

func NewPGRepository(db *sql.DB) (*PGRepository, error) {
	return &PGRepository{
		db: db,
	}, nil
}

// func (repo *PGRepository) SaveURL(url string) (id string, exists bool, err error) {

// }

// func (repo *PGRepository) RetrieveURL(id string) (url string, err error) {

// }

func (repo *PGRepository) CheckStatus(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		if err := repo.db.PingContext(ctx); err != nil {
			return err
		}
	}

	return nil
}
