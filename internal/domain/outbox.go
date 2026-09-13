package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// OutboxEvent is a to-be-published broker event persisted
type OutboxEvent struct {
	ID        int64
	Key       string // kafka message key, avatar id
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

// NewOutboxEvent marshals payload once, at enqueue time
func NewOutboxEvent(key, eventType string, payload any) (OutboxEvent, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return OutboxEvent{}, fmt.Errorf("marshal %s payload: %w", eventType, err)
	}
	return OutboxEvent{Key: key, Type: eventType, Payload: raw}, nil
}
