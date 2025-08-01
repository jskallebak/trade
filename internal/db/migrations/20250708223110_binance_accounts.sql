-- +goose Up
-- +goose StatementBegin
CREATE TABLE binance_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    api_secret VARCHAR(255) NOT NULL,
    base_url VARCHAR(255) DEFAULT 'https://api.binance.com',
    margin_enabled BOOLEAN DEFAULT false,  -- ← Add this line
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, name)
);
-- +goose StatementEnd

-- +goose Down  
-- +goose StatementBegin
DROP TABLE binance_accounts;
-- +goose StatementEnd
