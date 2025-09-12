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

-- name: GetLatestCompleteDayBalance :many
WITH latest_complete_day AS (
    SELECT DATE_TRUNC('day', MAX(recorded_at)) as complete_day
    FROM balance_history 
    WHERE DATE_TRUNC('day', recorded_at) < DATE_TRUNC('day', NOW())
),
day_balances AS (
    SELECT 
        bh.binance_account_id,
        bh.total_balance_usd,
        bh.recorded_at,
        ba.name as account_name,
        ba.user_id,
        ROW_NUMBER() OVER (
            PARTITION BY bh.binance_account_id 
            ORDER BY bh.recorded_at DESC
        ) as rn
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    CROSS JOIN latest_complete_day lcd
    WHERE DATE_TRUNC('day', bh.recorded_at) = lcd.complete_day
      AND ba.is_active = true
)
SELECT 
    binance_account_id,
    account_name,
    user_id,
    total_balance_usd,
    recorded_at
FROM day_balances 
WHERE rn = 1
ORDER BY user_id, account_name;

-- name: GetLatestCompleteDayBalanceByUser :many
WITH latest_complete_day AS (
    SELECT DATE_TRUNC('day', MAX(recorded_at)) as complete_day
    FROM balance_history 
    WHERE DATE_TRUNC('day', recorded_at) < DATE_TRUNC('day', NOW())
),
day_balances AS (
    SELECT 
        bh.binance_account_id,
        bh.total_balance_usd,
        bh.recorded_at,
        ba.name as account_name,
        ba.user_id,
        ROW_NUMBER() OVER (
            PARTITION BY bh.binance_account_id 
            ORDER BY bh.recorded_at DESC
        ) as rn
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    CROSS JOIN latest_complete_day lcd
    WHERE DATE_TRUNC('day', bh.recorded_at) = lcd.complete_day
      AND ba.is_active = true
      AND ba.user_id = $1
)
SELECT 
    binance_account_id,
    account_name,
    user_id,
    total_balance_usd,
    recorded_at
FROM day_balances 
WHERE rn = 1
ORDER BY account_name;

-- name: GetUserTotalBalanceLatestCompleteDay :one
WITH latest_complete_day AS (
    SELECT DATE_TRUNC('day', MAX(recorded_at)) as complete_day
    FROM balance_history 
    WHERE DATE_TRUNC('day', recorded_at) < DATE_TRUNC('day', NOW())
),
day_balances AS (
    SELECT 
        bh.binance_account_id,
        bh.total_balance_usd,
        ba.user_id,
        ROW_NUMBER() OVER (
            PARTITION BY bh.binance_account_id 
            ORDER BY bh.recorded_at DESC
        ) as rn
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    CROSS JOIN latest_complete_day lcd
    WHERE DATE_TRUNC('day', bh.recorded_at) = lcd.complete_day
      AND ba.is_active = true
      AND ba.user_id = $1
)
SELECT 
    COALESCE(SUM(total_balance_usd), 0) as total_balance_usd
FROM day_balances 
WHERE rn = 1;

-- name: GetUserTotalBalanceLatestCompleteMonth :one
WITH target_month AS (
    SELECT 
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM balance_history 
                WHERE DATE_TRUNC('month', recorded_at) < DATE_TRUNC('month', NOW())
            )
            THEN DATE_TRUNC('month', NOW()) - INTERVAL '1 month'  -- Last complete month
            ELSE DATE_TRUNC('month', NOW())                       -- Current month as fallback
        END as month_start
),
month_balances AS (
    SELECT 
        bh.binance_account_id, 
        bh.total_balance_usd,
        ba.user_id,
        ROW_NUMBER() OVER (
            PARTITION BY bh.binance_account_id
            ORDER BY bh.recorded_at ASC  -- EARLIEST instance from the target month
        ) as rn
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    CROSS JOIN target_month tm
    WHERE DATE_TRUNC('month', bh.recorded_at) = tm.month_start
        AND ba.is_active = true
        AND ba.user_id = $1
)
SELECT
    COALESCE(SUM(total_balance_usd), 0) as total_balance_usd
FROM month_balances
WHERE rn = 1;

-- name: GetUserTotalBalanceEarliestInYear :one
WITH target_year AS (
    SELECT 
        COALESCE(
            (SELECT MAX(DATE_TRUNC('year', recorded_at)) 
             FROM balance_history 
             WHERE DATE_TRUNC('year', recorded_at) < DATE_TRUNC('year', NOW())),
            DATE_TRUNC('year', NOW())
        ) as year_start
),
earliest_balances AS (
    SELECT DISTINCT ON (bh.binance_account_id) 
        bh.binance_account_id,
        bh.total_balance_usd
    FROM balance_history bh
    JOIN binance_accounts ba ON bh.binance_account_id = ba.id
    CROSS JOIN target_year ty
    WHERE DATE_TRUNC('year', bh.recorded_at) = ty.year_start
        AND ba.is_active = true
        AND ba.user_id = $1
    ORDER BY bh.binance_account_id, bh.recorded_at ASC
)
SELECT
    COALESCE(SUM(total_balance_usd), 0) as total_balance_usd
FROM earliest_balances;
