-- name: CreateBalanceRecord :one
INSERT INTO balance_history (binance_account_id, total_balance_usd, recorded_at)
VALUES ($1, $2, $3)
RETURNING id, binance_account_id, total_balance_usd, recorded_at;

-- name: GetLatestBalanceByAccount :one
SELECT id, binance_account_id, total_balance_usd, recorded_at
FROM balance_history
WHERE binance_account_id = $1
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetAccountBalanceHistory :many
SELECT id, binance_account_id, total_balance_usd, recorded_at
FROM balance_history
WHERE binance_account_id = $1 
AND recorded_at >= $2
ORDER BY recorded_at DESC
LIMIT $3;

-- name: GetUserTotalBalance :one
SELECT COALESCE(SUM(bh.total_balance_usd), 0) as total_balance_usd
FROM (
    SELECT DISTINCT ON (binance_account_id) total_balance_usd
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    WHERE ba.user_id = $1 AND ba.is_active = true
    ORDER BY binance_account_id, recorded_at DESC
) latest_balances;

-- name: DeleteOldBalanceRecords :exec
DELETE FROM balance_history
WHERE recorded_at < $1;
