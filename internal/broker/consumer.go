package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EventHandler processes one decoded event
type EventHandler func(ctx context.Context, envelope domain.EventEnvelope) error

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
			// Undecodable garbage. Log and move on
			c.logger.Error("skipping malformed event",
				zap.String("key", string(message.Key)),
				zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.Error(err),
			)
		} else if err := handle(ctx, envelope); err != nil {
			c.logger.Error("event handler gave up on message",
				zap.String("type", envelope.Type),
				zap.String("key", string(message.Key)),
				zap.Error(err),
			)
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("commit offset: %w", err)
		}
	}
}
