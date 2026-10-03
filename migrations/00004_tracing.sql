-- +goose Up
ALTER TABLE outbox_events ADD COLUMN trace_context JSONB;
-- +goose Down
ALTER TABLE outbox_events DROP COLUMN trace_context;