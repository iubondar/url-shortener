-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN DEFAULT false;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE urls DROP COLUMN is_deleted;
