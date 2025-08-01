-- +goose Up
-- +goose StatementBegin
ALTER TABLE bots ADD COLUMN binance_account_id INTEGER UNIQUE REFERENCES binance_accounts(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE bots DROP COLUMN binance_account_id;
-- +goose StatementEnd
