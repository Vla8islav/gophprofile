// Package broker implements domain.EventPublisher on top of Kafka.
package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer  *kafka.Writer
	brokers []string
	topic   string
}

// topicPartitions bounds the worker groups parallelism
const topicPartitions = 3

// NewKafkaPublisher builds the producer and attempts to create a topic
func NewKafkaPublisher(ctx context.Context, brokers []string, topic string) *KafkaPublisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 10 * time.Millisecond,
	}
	publisher := &KafkaPublisher{writer: writer, brokers: brokers, topic: topic}
	_ = publisher.EnsureTopic(ctx)
	return publisher
}

// EnsureTopic idempotently creates the topic
func (p *KafkaPublisher) EnsureTopic(ctx context.Context) error {
	dialer := &kafka.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer func() { _ = conn.Close() }()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("find controller: %w", err)
	}
	controllerConn, err := dialer.DialContext(ctx, "tcp",
		net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer func() { _ = controllerConn.Close() }()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             p.topic,
		NumPartitions:     topicPartitions,
		ReplicationFactor: 1,
	})
	if err != nil && !errors.Is(err, kafka.TopicAlreadyExists) {
		return fmt.Errorf("create topic %s: %w", p.topic, err)
	}
	return nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, key string, eventType string, payload any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", eventType, err)
	}
	envelope, err := json.Marshal(domain.EventEnvelope{
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		Payload:    payloadJSON,
	})
	if err != nil {
		return fmt.Errorf("marshal %s envelope: %w", eventType, err)
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: envelope,
	}

	for attempt := 1; ; attempt++ {
		err = p.writer.WriteMessages(ctx, message)
		if err == nil {
			return nil
		}
		if attempt >= publishMaxAttempts || !isRetriablePublishError(err) {
			return fmt.Errorf("publish %s to %s: %w", eventType, p.topic, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * publishRetryDelay):
		}
	}
}

const (
	publishMaxAttempts = 5
	publishRetryDelay  = 200 * time.Millisecond
)

func isRetriablePublishError(err error) bool {
	if errors.Is(err, kafka.UnknownTopicOrPartition) || errors.Is(err, kafka.LeaderNotAvailable) {
		return true
	}
	var writeErrs kafka.WriteErrors
	if errors.As(err, &writeErrs) {
		for _, e := range writeErrs {
			if isRetriablePublishError(e) {
				return true
			}
		}
	}
	return false
}

// Ping fetches topic metadata from the first reachable broker
func (p *KafkaPublisher) Ping(ctx context.Context) error {
	dialer := &kafka.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("broker unreachable: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Brokers(); err != nil {
		return fmt.Errorf("broker metadata: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
