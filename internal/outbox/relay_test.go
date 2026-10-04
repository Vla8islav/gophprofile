package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func event(id int64, key string) domain.OutboxEvent {
	return domain.OutboxEvent{
		ID: id, Key: key, Type: domain.EventTypeAvatarUploaded,
		Payload: json.RawMessage(`{"avatar_id":"` + key + `"}`),
	}
}

func feed(evs ...domain.OutboxEvent) func(context.Context, int, func(context.Context, domain.OutboxEvent) error) (int, error) {
	return func(ctx context.Context, _ int, handle func(context.Context, domain.OutboxEvent) error) (int, error) {
		n := 0
		for _, ev := range evs {
			if err := handle(ctx, ev); err != nil {
				return n, nil
			}
			n++
		}
		return n, nil
	}
}

func TestRelay_DrainsUntilEmpty(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	gomock.InOrder(
		repo.EXPECT().ProcessUnsentOutboxEvents(gomock.Any(), batchSize, gomock.Any()).
			DoAndReturn(feed(event(1, "av-1"), event(2, "av-2"))),
		repo.EXPECT().ProcessUnsentOutboxEvents(gomock.Any(), batchSize, gomock.Any()).
			Return(0, nil),
	)
	gomock.InOrder(
		publisher.EXPECT().Publish(gomock.Any(), "av-1",
			domain.EventTypeAvatarUploaded, gomock.Any()).Return(nil),
		publisher.EXPECT().Publish(gomock.Any(), "av-2",
			domain.EventTypeAvatarUploaded, gomock.Any()).Return(nil),
	)

	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
}

func TestRelay_PublishFailureStopsBatch(t *testing.T) {
	// no t.Parallel: asserts a global counter
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	repo.EXPECT().ProcessUnsentOutboxEvents(gomock.Any(), batchSize, gomock.Any()).
		DoAndReturn(feed(event(1, "av-1"), event(2, "av-2")))
	publisher.EXPECT().Publish(gomock.Any(), "av-1", gomock.Any(), gomock.Any()).
		Return(errors.New("kafka down"))
	// no av-2 expectation: if the relay publishes it anyway, ctrl.Finish fails the test

	before := testutil.ToFloat64(outboxPublishFailures)
	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
	if d := testutil.ToFloat64(outboxPublishFailures) - before; d != 1 {
		t.Errorf("publish_failures delta = %v, want 1", d)
	}
}

func TestRelay_RepoErrorStops(t *testing.T) {
	// no t.Parallel: asserts a global counter
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	repo.EXPECT().ProcessUnsentOutboxEvents(gomock.Any(), batchSize, gomock.Any()).
		Return(0, errors.New("db down"))
	// zero publisher expectations: repo failure must not reach the broker

	before := testutil.ToFloat64(outboxPublishFailures)
	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
	if d := testutil.ToFloat64(outboxPublishFailures) - before; d != 0 {
		t.Errorf("publish_failures delta = %v, want 0 (repo failure misreported as broker trouble)", d)
	}
}
