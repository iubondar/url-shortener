-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
	short_url VARCHAR(10) UNIQUE NOT NULL,
	original_url VARCHAR(2048) UNIQUE NOT NULL);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_index ON urls (short_url);

CREATE UNIQUE INDEX IF NOT EXISTS original_url_index ON urls (original_url);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS original_url_index;

DROP INDEX IF EXISTS short_url_index;

DROP TABLE IF EXISTS urls;
