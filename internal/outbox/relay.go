// Package outbox ships events from the transactional outbox table to the broker
package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

const (
	pollInterval = time.Second
	batchSize    = 100
	batchTimeout = 30 * time.Second
)

// Repository is the slice of the storage layer
type Repository interface {
	ProcessUnsentOutboxEvents(ctx context.Context, limit int,
		handle func(context.Context, domain.OutboxEvent) error) (processed, fetched int, err error)
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
		batchCtx, cancel := context.WithTimeout(ctx, batchTimeout)
		processed, fetched, err := r.repository.ProcessUnsentOutboxEvents(batchCtx, batchSize, r.relayOne)
		cancel()
		if err != nil {
			r.logger.Error("outbox: failed to process unsent events", zap.Error(err))
			return
		}
		if processed == 0 {
			if fetched == 0 {
				outboxOldestPendingAge.Set(0) // queue empty
			}
			return // fetched > 0; processing went wrong
		}
	}
}

func (r *Relay) relayOne(ctx context.Context, event domain.OutboxEvent) error {
	outboxOldestPendingAge.Set(time.Since(event.CreatedAt).Seconds())
	carrier := propagation.MapCarrier{}
	if len(event.TraceContext) > 0 {
		if err := json.Unmarshal(event.TraceContext, &carrier); err != nil {
			r.logger.Warn("outbox: malformed trace context, publishing without trace",
				zap.Int64("event_id", event.ID), zap.Error(err))
		}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := tracer.Start(ctx, "outbox.publish")
	defer span.End()
	ctx = logging.Into(ctx, logging.WithTrace(ctx, r.logger))

	if err := r.publisher.Publish(ctx, event.Key, event.Type, event.Payload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		outboxPublishFailures.Inc()
		logging.From(ctx).Warn("outbox: publish failed, will retry next tick",
			zap.Int64("event_id", event.ID),
			zap.String("type", event.Type),
			zap.Error(err),
		)
		return err
	}
	outboxPublished.Inc()
	return nil
}
