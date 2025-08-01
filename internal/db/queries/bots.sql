-- name: CreateBot :one
INSERT INTO bots (user_id, name, strategy, initial_holding, binance_account_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, binance_account_id, created_at, updated_at;


-- name: GetUserBots :many
SELECT id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, binance_account_id, created_at, updated_at
FROM bots
WHERE user_id = $1;

-- name: GetBot :one
SELECT id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, created_at, updated_at
FROM bots
WHERE id = $1 AND user_id = $2;

-- name: UpdateBotStatus :one
UPDATE bots
SET 
    status = $3,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, created_at, updated_at;

-- name: DeleteBot :exec
DELETE FROM bots
WHERE id = $1 AND user_id = $2;

-- name: UpdateBot :one
UPDATE bots
SET
    name = $3,
    strategy = $4,
    initial_holding = $5,
    binance_account_id = $6,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, binance_account_id, created_at, updated_at;

-- name: GetUserBotsWithAccounts :many
SELECT 
    b.id, b.user_id, b.name, b.strategy, b.status, b.win_rate, b.profit_factor, b.trades, 
    b.initial_holding, b.holding, b.binance_account_id, b.created_at, b.updated_at,
    ba.name as account_name
FROM bots b
LEFT JOIN binance_accounts ba ON b.binance_account_id = ba.id
WHERE b.user_id = $1;

