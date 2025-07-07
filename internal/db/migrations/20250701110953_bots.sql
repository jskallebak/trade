-- +goose Up
-- +goose StatementBegin
CREATE TABLE bots (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    strategy VARCHAR(255) NOT NULL,
    status VARCHAR(255) DEFAULT 'STOPPED',
    win_rate DECIMAL(5,2) DEFAULT 0.0,
    profit_factor DECIMAL(10,2) DEFAULT 0.0,
    trades INT DEFAULT 0,
    initial_holding DECIMAL(15,2) DEFAULT 0.0,
    holding DECIMAL(15,2) DEFAULT 0.0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT check_win_rate CHECK (win_rate >= 0 AND win_rate <= 100),
    CONSTRAINT check_status CHECK (status IN ('STOPPED', 'RUNNING', 'PAUSED', 'ERROR')),
    CONSTRAINT check_profit_factor CHECK (profit_factor >= 0),
    CONSTRAINT unique_user_bot_name UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE

);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE bots
-- +goose StatementEnd
