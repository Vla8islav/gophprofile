package domain

import (
	"context"
	"encoding/json"
	"time"
)

const (
	EventTypeAvatarUploaded = "avatar.uploaded"
	EventTypeAvatarDeleted  = "avatar.deleted"
)

// AvatarUploadEvent tells the worker a new original landed in S3 and needs
// thumbnails.
type AvatarUploadEvent struct {
	AvatarID string `json:"avatar_id"`
	UserID   int64  `json:"user_id"`
	S3Key    string `json:"s3_key"`
}

// AvatarDeleteEvent tells the worker which S3 objects to remove. It carries
// the keys explicitly because by the time the worker runs, the DB row is
// already soft-deleted and invisible to normal reads.
type AvatarDeleteEvent struct {
	AvatarID string   `json:"avatar_id"`
	S3Keys   []string `json:"s3_keys"`
}

// EventEnvelope is the wire format on the avatar-events topic. All event
// types share one topic, keyed by avatar id, so Kafka preserves per-avatar
// ordering (an upload is never processed after its own delete).
type EventEnvelope struct {
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// EventPublisher abstracts the message broker's producer side.
// key groups related events for ordering (we always pass the avatar id).
type EventPublisher interface {
	Publish(ctx context.Context, key string, eventType string, payload any) error
	Ping(ctx context.Context) error
	Close() error
}
