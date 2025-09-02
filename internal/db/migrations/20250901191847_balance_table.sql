-- +goose Up
-- +goose StatementBegin
CREATE TABLE balance_history (
    id SERIAL PRIMARY KEY,
    binance_account_id INTEGER NOT NULL REFERENCES binance_accounts(id) ON DELETE CASCADE,
    total_balance_usd DECIMAL(15,2) NOT NULL DEFAULT 0.0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes for better performance
CREATE INDEX idx_balance_history_account_id ON balance_history(binance_account_id);
CREATE INDEX idx_balance_history_recorded_at ON balance_history(recorded_at);
CREATE INDEX idx_balance_history_account_time ON balance_history(binance_account_id, recorded_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balance_history;
-- +goose StatementEnd
