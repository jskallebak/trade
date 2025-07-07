-- name: CreateBot :one
INSERT INTO bots (user_id, name, strategy, initial_holding)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, created_at, updated_at;

-- name: GetUserBots :many
SELECT id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, created_at, updated_at
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
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, strategy, status, win_rate, profit_factor, trades, initial_holding, holding, created_at, updated_at;

