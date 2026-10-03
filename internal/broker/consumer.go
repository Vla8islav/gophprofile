package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/logging"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EventHandler processes one decoded event
type EventHandler func(ctx context.Context, envelope domain.EventEnvelope) error

const (
	handlerBackoffBase = time.Second
	handlerBackoffMax  = 30 * time.Second
)

type KafkaConsumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

func NewKafkaConsumer(brokers []string, topic, groupID string, logger *zap.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
		// speeds up re-join
		JoinGroupBackoff: time.Second,
	})
	return &KafkaConsumer{reader: reader, logger: logger}
}

// Run consumes until ctx is cancelled
func (c *KafkaConsumer) Run(ctx context.Context, handle EventHandler) error {
	defer func() { _ = c.reader.Close() }()

	for {
		message, err := c.reader.FetchMessage(ctx)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil // graceful shutdown
		}
		if err != nil {
			return fmt.Errorf("fetch message: %w", err)
		}

		var envelope domain.EventEnvelope
		if err := json.Unmarshal(message.Value, &envelope); err != nil {
			// garbage
			consumerEvents.WithLabelValues("malformed").Inc()
			c.logger.Error("skipping malformed event",
				zap.String("key", string(message.Key)),
				zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.Error(err),
			)

		} else {
			msgCtx := otel.GetTextMapPropagator().Extract(ctx, kafkaHeaderCarrier{&message.Headers})
			msgCtx, span := tracer.Start(msgCtx, "consume "+envelope.Type,
				trace.WithSpanKind(trace.SpanKindConsumer))
			msgCtx = logging.Into(msgCtx, logging.WithTrace(msgCtx, c.logger))
			err := c.handleWithRetry(msgCtx, handle, envelope, message)
			span.End()
			if err != nil {
				return nil // shutdown mid-retry
			}
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("commit offset: %w", err)
		}
	}
}

// handleWithRetry drives one message to a terminal state: success, permanent failure (drop), or shutdown
func (c *KafkaConsumer) handleWithRetry(ctx context.Context, handle EventHandler, envelope domain.EventEnvelope, message kafka.Message) error {
	for attempt := 1; ; attempt++ {
		err := handle(ctx, envelope)
		if err == nil {
			consumerEvents.WithLabelValues("ok").Inc()
			return nil
		}
		if IsPermanent(err) {
			consumerEvents.WithLabelValues("permanent").Inc()
			logging.From(ctx).Error("dropping event after permanent failure",
				zap.String("type", envelope.Type),
				zap.String("key", string(message.Key)),
				zap.Error(err),
			)
			return nil
		}

		backoff := handlerBackoffBase << (attempt - 1)
		if backoff > handlerBackoffMax || backoff <= 0 { // <=0 guards shift overflow
			backoff = handlerBackoffMax
		}
		logging.From(ctx).Warn("transient failure, will retry same message",
			zap.String("type", envelope.Type),
			zap.String("key", string(message.Key)),
			zap.Int("attempt", attempt),
			zap.Duration("backoff", backoff),
			zap.Error(err),
		)
		consumerEvents.WithLabelValues("retry").Inc()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
}
