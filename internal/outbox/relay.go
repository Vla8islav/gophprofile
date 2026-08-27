// Package outbox ships events from the transactional outbox table to the broker
package outbox

import (
	"context"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.uber.org/zap"
)

const (
	pollInterval = time.Second
	batchSize    = 100
)

// Repository is the slice of the storage layer
type Repository interface {
	UnsentOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkOutboxEventSent(ctx context.Context, eventID int64) error
}

// Publisher is the producing slice of domain.EventPublisher.
type Publisher interface {
	Publish(ctx context.Context, key string, eventType string, payload any) error
}

type Relay struct {
	repository Repository
	publisher  Publisher
	logger     *zap.Logger
}

func NewRelay(repository Repository, publisher Publisher, logger *zap.Logger) *Relay {
	return &Relay{repository: repository, publisher: publisher, logger: logger}
}

// Run polls until ctx is cancelled. Failures are logged and retried
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.drain(ctx)
		}
	}
}

// drain publishes unsent events oldest-first until the table is empty an error halts it
func (r *Relay) drain(ctx context.Context) {
	for {
		events, err := r.repository.UnsentOutboxEvents(ctx, batchSize)
		if err != nil {
			r.logger.Error("outbox: failed to list unsent events", zap.Error(err))
			return
		}
		if len(events) == 0 {
			return
		}

		for _, event := range events {
			if err := r.publisher.Publish(ctx, event.Key, event.Type, event.Payload); err != nil {
				r.logger.Warn("outbox: publish failed, will retry next tick",
					zap.Int64("event_id", event.ID),
					zap.String("type", event.Type),
					zap.Error(err),
				)
				return
			}
			if err := r.repository.MarkOutboxEventSent(ctx, event.ID); err != nil {
				// The event WAS published
				r.logger.Error("outbox: failed to mark event sent",
					zap.Int64("event_id", event.ID),
					zap.Error(err),
				)
				return
			}
		}
	}
}
