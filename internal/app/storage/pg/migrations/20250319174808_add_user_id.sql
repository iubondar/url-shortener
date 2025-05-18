-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE urls ADD COLUMN user_id uuid;

CREATE INDEX IF NOT EXISTS user_id_index ON urls (user_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS user_id_index;

ALTER TABLE urls DROP COLUMN user_id;
