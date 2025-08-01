-- name: CreateBinanceAccount :one
INSERT INTO binance_accounts (user_id, name, api_key, api_secret, base_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, api_key, base_url, is_active, created_at, updated_at;

-- name: GetUserBinanceAccounts :many  
SELECT id, user_id, name, api_key, base_url, is_active, created_at, updated_at
FROM binance_accounts 
WHERE user_id = $1 AND is_active = true;

-- name: GetBinanceAccount :one
SELECT id, user_id, name, api_key, api_secret, base_url, is_active
FROM binance_accounts
WHERE id = $1 AND user_id = $2 AND is_active = true;

-- name: UpdateBinanceAccount :one
UPDATE binance_accounts
SET name = $3, api_key = $4, api_secret = $5, base_url = $6, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, api_key, base_url, is_active, created_at, updated_at;

-- name: DeleteBinanceAccount :exec
UPDATE binance_accounts 
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: GetInactiveBinanceAccount :one
SELECT id, user_id, name, api_key, api_secret, base_url, is_active, created_at, updated_at
FROM binance_accounts
WHERE user_id = $1 AND name = $2 AND is_active = false;

-- name: ReactivateBinanceAccount :one
UPDATE binance_accounts
SET 
    api_key = $3,
    api_secret = $4,
    base_url = $5,
    is_active = true,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, api_key, base_url, is_active, created_at, updated_at;

-- name: UpdateBinanceAccountInfo :one
UPDATE binance_accounts
SET 
    name = $3,
    base_url = $4,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND is_active = true
RETURNING id, user_id, name, api_key, base_url, is_active, created_at, updated_at;

-- name: GetUserBinanceAccountsWithStatus :many
SELECT 
    ba.id, ba.user_id, ba.name, ba.api_key, ba.base_url, ba.margin_enabled, ba.is_active, ba.created_at, ba.updated_at,
    CASE 
        WHEN b.id IS NOT NULL THEN true 
        ELSE false 
    END as account_active
FROM binance_accounts ba
LEFT JOIN bots b ON ba.id = b.binance_account_id
WHERE ba.user_id = $1 AND ba.is_active = true;
