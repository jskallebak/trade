-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN password_hash;
-- +goose StatementEnd
