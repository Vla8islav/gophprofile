-- +goose Up
-- Outbox: events are inserted in the same transaction as the state change
-- they announce. A broker outage delays events instead of losing them.
CREATE TABLE outbox_events (
                               id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                               event_key  VARCHAR(64)  NOT NULL,
                               event_type VARCHAR(64)  NOT NULL,
                               payload    JSONB        NOT NULL,
                               created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
                               sent_at    TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_unsent
    ON outbox_events (id)
    WHERE sent_at IS NULL;

-- +goose Down
DROP TABLE outbox_events;