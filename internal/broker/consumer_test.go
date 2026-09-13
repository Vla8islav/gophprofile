package broker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestKafkaConsumer_ReceivesPublishedEvents(t *testing.T) {
	publisher, seedBroker := startPublisher(t)
	ctx := context.Background()

	require.NoError(t, publisher.Publish(ctx, "av-1", domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: "av-1"}))
	require.NoError(t, publisher.Publish(ctx, "av-2", domain.EventTypeAvatarDeleted,
		domain.AvatarDeleteEvent{AvatarID: "av-2"}))

	var mu sync.Mutex
	received := map[string]int{}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	consumer := NewKafkaConsumer([]string{seedBroker}, "avatar-events-test", "consumer-test", zap.NewNop())
	done := make(chan error, 1)
	go func() {
		done <- consumer.Run(runCtx, func(_ context.Context, envelope domain.EventEnvelope) error {
			mu.Lock()
			received[envelope.Type]++
			mu.Unlock()
			return nil
		})
	}()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return received[domain.EventTypeAvatarUploaded] == 1 &&
			received[domain.EventTypeAvatarDeleted] == 1
	}, 30*time.Second, 100*time.Millisecond)

	cancel()
	require.NoError(t, <-done)
}

func TestKafkaConsumer_PermanentFailureDoesNotWedgePartition(t *testing.T) {
	publisher, seedBroker := startPublisher(t)
	ctx := context.Background()

	// Same key: same partition — order is guaranteed.
	require.NoError(t, publisher.Publish(ctx, "poison", domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: "poison"}))
	require.NoError(t, publisher.Publish(ctx, "poison", domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: "fine"}))

	var mu sync.Mutex
	var order []string

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	consumer := NewKafkaConsumer([]string{seedBroker}, "avatar-events-test", "poison-test", zap.NewNop())
	go func() {
		_ = consumer.Run(runCtx, func(_ context.Context, envelope domain.EventEnvelope) error {
			var event domain.AvatarUploadEvent
			require.NoError(t, json.Unmarshal(envelope.Payload, &event))
			mu.Lock()
			order = append(order, event.AvatarID)
			mu.Unlock()
			if event.AvatarID == "poison" {
				return Permanent(errors.New("corrupt image"))
			}
			return nil
		})
	}()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(order) == 2 && order[0] == "poison" && order[1] == "fine"
	}, 30*time.Second, 100*time.Millisecond)
}

func TestKafkaConsumer_TransientFailureRetriesSameMessage(t *testing.T) {
	publisher, seedBroker := startPublisher(t)
	ctx := context.Background()

	require.NoError(t, publisher.Publish(ctx, "flaky", domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: "flaky"}))
	require.NoError(t, publisher.Publish(ctx, "flaky", domain.EventTypeAvatarUploaded,
		domain.AvatarUploadEvent{AvatarID: "after"}))

	var mu sync.Mutex
	attempts := map[string]int{}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	consumer := NewKafkaConsumer([]string{seedBroker}, "avatar-events-test", "flaky-test", zap.NewNop())
	go func() {
		_ = consumer.Run(runCtx, func(_ context.Context, envelope domain.EventEnvelope) error {
			var event domain.AvatarUploadEvent
			require.NoError(t, json.Unmarshal(envelope.Payload, &event))
			mu.Lock()
			attempts[event.AvatarID]++
			count := attempts[event.AvatarID]
			mu.Unlock()
			// "the S3 outage": fails twice, works on the third attempt
			if event.AvatarID == "flaky" && count < 3 {
				return errors.New("minio down")
			}
			return nil
		})
	}()

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		// the flaky event was attempted exactly 3 times
		return attempts["flaky"] == 3 && attempts["after"] == 1
	}, 30*time.Second, 100*time.Millisecond)
}
