package broker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func startPublisher(t *testing.T) (*KafkaPublisher, string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcredpanda.Run(ctx, "redpandadata/redpanda:v24.3.6")
	require.NoError(t, err, "failed to start redpanda container")
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	seedBroker, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)

	publisher := NewKafkaPublisher(ctx, []string{seedBroker}, "avatar-events-test")
	require.NoError(t, publisher.EnsureTopic(ctx))
	t.Cleanup(func() { _ = publisher.Close() })
	return publisher, seedBroker
}

func TestKafkaPublisher_PublishAndConsume(t *testing.T) {
	publisher, seedBroker := startPublisher(t)
	ctx := context.Background()

	event := domain.AvatarUploadEvent{
		AvatarID: "7c9e6679-7425-40de-944b-e07fc1f90ae7",
		UserID:   42,
		S3Key:    "avatars/7c9e6679-7425-40de-944b-e07fc1f90ae7/original",
	}
	require.NoError(t, publisher.Publish(ctx, event.AvatarID, domain.EventTypeAvatarUploaded, event))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{seedBroker},
		Topic:   "avatar-events-test",
		GroupID: "test-consumer",
	})
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	msg, err := reader.ReadMessage(readCtx)
	require.NoError(t, err)

	// The key carries the avatar id (this is what gives per-avatar ordering).
	require.Equal(t, event.AvatarID, string(msg.Key))

	var envelope domain.EventEnvelope
	require.NoError(t, json.Unmarshal(msg.Value, &envelope))
	require.Equal(t, domain.EventTypeAvatarUploaded, envelope.Type)
	require.False(t, envelope.OccurredAt.IsZero())

	var decoded domain.AvatarUploadEvent
	require.NoError(t, json.Unmarshal(envelope.Payload, &decoded))
	require.Equal(t, event, decoded)
}

func TestKafkaPublisher_SameKeySamePartition(t *testing.T) {
	publisher, seedBroker := startPublisher(t)
	ctx := context.Background()

	// Publish an upload and a delete for the same avatar; they must land in
	// the same partition in order, or the worker could process them swapped.
	avatarID := "0b2a4f0a-9c58-4d8e-9d5c-111111111111"
	require.NoError(t, publisher.Publish(ctx, avatarID, domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: avatarID}))
	require.NoError(t, publisher.Publish(ctx, avatarID, domain.EventTypeAvatarDeleted,
		domain.AvatarDeleteEvent{AvatarID: avatarID}))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{seedBroker},
		Topic:   "avatar-events-test",
		GroupID: "ordering-consumer",
	})
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	first, err := reader.ReadMessage(readCtx)
	require.NoError(t, err)
	second, err := reader.ReadMessage(readCtx)
	require.NoError(t, err)

	require.Equal(t, first.Partition, second.Partition)

	var env1, env2 domain.EventEnvelope
	require.NoError(t, json.Unmarshal(first.Value, &env1))
	require.NoError(t, json.Unmarshal(second.Value, &env2))
	require.Equal(t, domain.EventTypeAvatarUploaded, env1.Type)
	require.Equal(t, domain.EventTypeAvatarDeleted, env2.Type)
}

func TestKafkaPublisher_Ping(t *testing.T) {
	publisher, _ := startPublisher(t)
	require.NoError(t, publisher.Ping(context.Background()))
}

func TestKafkaPublisher_PingUnreachableBroker(t *testing.T) {
	publisher := NewKafkaPublisher(context.Background(), []string{"localhost:1"}, "nope")
	defer publisher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.Error(t, publisher.Ping(ctx))
}
