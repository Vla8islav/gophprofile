package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func event(id int64, key string) domain.OutboxEvent {
	return domain.OutboxEvent{
		ID: id, Key: key, Type: domain.EventTypeAvatarUploaded,
		Payload: json.RawMessage(`{"avatar_id":"` + key + `"}`),
	}
}

func TestRelay_PublishesAndMarksInOrder(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	// first pass returns two events, second pass returns none (drained)
	gomock.InOrder(
		repo.EXPECT().UnsentOutboxEvents(gomock.Any(), batchSize).
			Return([]domain.OutboxEvent{event(1, "av-1"), event(2, "av-2")}, nil),
		publisher.EXPECT().Publish(gomock.Any(), "av-1", domain.EventTypeAvatarUploaded, gomock.Any()).Return(nil),
		repo.EXPECT().MarkOutboxEventSent(gomock.Any(), int64(1)).Return(nil),
		publisher.EXPECT().Publish(gomock.Any(), "av-2", domain.EventTypeAvatarUploaded, gomock.Any()).Return(nil),
		repo.EXPECT().MarkOutboxEventSent(gomock.Any(), int64(2)).Return(nil),
		repo.EXPECT().UnsentOutboxEvents(gomock.Any(), batchSize).
			Return(nil, nil),
	)

	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
}

func TestRelay_PublishFailureStopsPassWithoutMarking(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	repo.EXPECT().UnsentOutboxEvents(gomock.Any(), batchSize).
		Return([]domain.OutboxEvent{event(1, "av-1"), event(2, "av-2")}, nil)
	publisher.EXPECT().
		Publish(gomock.Any(), "av-1", gomock.Any(), gomock.Any()).
		Return(errors.New("kafka down"))

	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
}

// A failed mark stops the pass too; the event will be re-published next tick
func TestRelay_MarkFailureStopsPass(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)

	repo.EXPECT().UnsentOutboxEvents(gomock.Any(), batchSize).
		Return([]domain.OutboxEvent{event(1, "av-1"), event(2, "av-2")}, nil)
	publisher.EXPECT().Publish(gomock.Any(), "av-1", gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().MarkOutboxEventSent(gomock.Any(), int64(1)).
		Return(errors.New("db hiccup"))

	NewRelay(repo, publisher, zap.NewNop()).drain(context.Background())
}
