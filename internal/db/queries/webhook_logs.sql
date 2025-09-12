-- name: CreateWebhookLog :one
INSERT INTO webhook_logs (
    webhook_source, event_type, method, url_path, headers, query_params, 
    request_body, response_status, response_body, ip_address, user_agent, 
    processing_time_ms, error_message, is_successful
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING id, webhook_source, event_type, created_at;

-- name: GetWebhookLog :one
SELECT id, webhook_source, event_type, method, url_path, headers, query_params,
       request_body, response_status, response_body, ip_address, user_agent,
       processing_time_ms, error_message, is_successful, created_at
FROM webhook_logs
WHERE id = $1;

-- name: GetWebhookLogs :many
SELECT id, webhook_source, event_type, method, url_path, response_status,
       ip_address, processing_time_ms, error_message, is_successful, created_at
FROM webhook_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetWebhookLogsBySource :many
SELECT id, webhook_source, event_type, method, url_path, response_status,
       ip_address, processing_time_ms, error_message, is_successful, created_at
FROM webhook_logs
WHERE webhook_source = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFailedWebhookLogs :many
SELECT id, webhook_source, event_type, method, url_path, response_status,
       ip_address, processing_time_ms, error_message, is_successful, created_at
FROM webhook_logs
WHERE is_successful = false
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetWebhookLogsByDateRange :many
SELECT id, webhook_source, event_type, method, url_path, response_status,
       ip_address, processing_time_ms, error_message, is_successful, created_at
FROM webhook_logs
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetWebhookStatsbySource :many
SELECT webhook_source,
       COUNT(*) as total_requests,
       COUNT(*) FILTER (WHERE is_successful = true) as successful_requests,
       COUNT(*) FILTER (WHERE is_successful = false) as failed_requests,
       AVG(processing_time_ms) as avg_processing_time_ms,
       MAX(created_at) as last_request_at
FROM webhook_logs
WHERE created_at >= $1
GROUP BY webhook_source
ORDER BY total_requests DESC;

-- name: DeleteOldWebhookLogs :exec
DELETE FROM webhook_logs
WHERE created_at < $1;
