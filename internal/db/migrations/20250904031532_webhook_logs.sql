-- internal/db/migrations/YYYYMMDDHHMMSS_webhook_logs.sql
-- Replace YYYYMMDDHHMMSS with actual timestamp

-- +goose Up
-- +goose StatementBegin
CREATE TABLE webhook_logs (
    id SERIAL PRIMARY KEY,
    webhook_source VARCHAR(100) NOT NULL,
    event_type VARCHAR(100),
    method VARCHAR(10) NOT NULL DEFAULT 'POST',
    url_path TEXT NOT NULL,
    headers JSONB,
    query_params JSONB,
    request_body JSONB,
    response_status INTEGER DEFAULT 200,
    response_body TEXT,
    ip_address INET,
    user_agent TEXT,
    processing_time_ms INTEGER DEFAULT 0,
    error_message TEXT,
    is_successful BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_source ON webhook_logs(webhook_source);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_event_type ON webhook_logs(event_type);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs(created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_successful ON webhook_logs(is_successful);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_source_created ON webhook_logs(webhook_source, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_webhook_logs_failed ON webhook_logs(created_at DESC) WHERE is_successful = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS webhook_logs;
-- +goose StatementEnd
