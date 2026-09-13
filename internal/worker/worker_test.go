package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/broker"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func envelopeFor(t *testing.T, eventType string, payload any) domain.EventEnvelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return domain.EventEnvelope{Type: eventType, Payload: raw}
}

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))))
	return buf.Bytes()
}

func TestHandleEvent_Uploaded_GeneratesThumbnails(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	avatar := &domain.Avatar{
		ID:               "av-1",
		S3Key:            "avatars/av-1/original",
		ProcessingStatus: domain.ProcessingStatusPending,
	}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "av-1").Return(avatar, nil)
	storage.EXPECT().Download(gomock.Any(), "avatars/av-1/original").
		Return(io.NopCloser(bytes.NewReader(pngBytes(t, 400, 200))), nil)
	storage.EXPECT().
		Upload(gomock.Any(), gomock.Any(), "image/jpeg", gomock.Any(), gomock.Any()).
		Times(2).
		Return(nil)
	repo.EXPECT().
		SetAvatarThumbnails(gomock.Any(), "av-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, keys map[string]string) error {
			require.Equal(t, map[string]string{
				"100x100": "thumbnails/av-1/100x100.jpg",
				"300x300": "thumbnails/av-1/300x300.jpg",
			}, keys)
			return nil
		})

	w := New(repo, storage, zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{AvatarID: "av-1"}))
	require.NoError(t, err)
}

func TestHandleEvent_Uploaded_SkipsAlreadyCompleted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "av-1").
		Return(&domain.Avatar{ID: "av-1", ProcessingStatus: domain.ProcessingStatusCompleted}, nil)
	// no storage expectations: nothing may be downloaded or uploaded

	w := New(repo, mocks.NewMockFileStorage(ctrl), zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{AvatarID: "av-1"}))
	require.NoError(t, err)
}

func TestHandleEvent_Uploaded_SkipsDeletedAvatar(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "gone").
		Return(nil, domain.ErrAvatarNotFound)

	w := New(repo, mocks.NewMockFileStorage(ctrl), zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{AvatarID: "gone"}))
	require.NoError(t, err)
}

// A storage outage is transient: the error is NOT permanent, the avatar is
// NOT marked failed (it stays pending), and the consumer will redeliver.
func TestHandleEvent_Uploaded_TransientStorageFailure(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	avatar := &domain.Avatar{ID: "av-1", S3Key: "avatars/av-1/original"}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "av-1").Return(avatar, nil)
	storage.EXPECT().Download(gomock.Any(), "avatars/av-1/original").
		Return(nil, errors.New("minio down"))
	// no SetAvatarProcessingStatus expectation: it must NOT be called

	w := New(repo, storage, zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{AvatarID: "av-1"}))
	require.Error(t, err)
	require.False(t, broker.IsPermanent(err))
}

// A corrupt original is permanent: marked failed once, error flagged so the
// consumer drops the message instead of retrying forever.
func TestHandleEvent_Uploaded_CorruptImageIsPermanent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	avatar := &domain.Avatar{ID: "av-1", S3Key: "avatars/av-1/original"}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "av-1").Return(avatar, nil)
	storage.EXPECT().Download(gomock.Any(), gomock.Any()).
		Return(io.NopCloser(bytes.NewReader([]byte("not an image"))), nil)
	repo.EXPECT().
		SetAvatarProcessingStatus(gomock.Any(), "av-1", domain.ProcessingStatusFailed).
		Return(nil)

	w := New(repo, storage, zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{AvatarID: "av-1"}))
	require.Error(t, err)
	require.True(t, broker.IsPermanent(err))
}

func TestHandleEvent_Deleted_RemovesAllKeys(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mocks.NewMockFileStorage(ctrl)
	storage.EXPECT().Delete(gomock.Any(), "avatars/av-1/original").Return(nil)
	storage.EXPECT().Delete(gomock.Any(), "thumbnails/av-1/100x100.jpg").Return(nil)

	w := New(mocks.NewMockGophprofileRepository(ctrl), storage, zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarDeleted, domain.AvatarDeleteEvent{
			AvatarID: "av-1",
			S3Keys:   []string{"avatars/av-1/original", "thumbnails/av-1/100x100.jpg"},
		}))
	require.NoError(t, err)
}

// Failed S3 deletes are transient: the consumer redelivers until the objects
// are actually gone — nothing leaks silently.
func TestHandleEvent_Deleted_FailureIsTransient(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mocks.NewMockFileStorage(ctrl)
	storage.EXPECT().Delete(gomock.Any(), "a").Return(errors.New("nope"))
	storage.EXPECT().Delete(gomock.Any(), "b").Return(nil)

	w := New(mocks.NewMockGophprofileRepository(ctrl), storage, zap.NewNop())
	err := w.HandleEvent(context.Background(),
		envelopeFor(t, domain.EventTypeAvatarDeleted, domain.AvatarDeleteEvent{
			AvatarID: "av-1", S3Keys: []string{"a", "b"},
		}))
	require.Error(t, err)
	require.False(t, broker.IsPermanent(err))
}

// A payload that doesn't decode can never succeed: permanent.
func TestHandleEvent_MalformedPayloadIsPermanent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := New(mocks.NewMockGophprofileRepository(ctrl), mocks.NewMockFileStorage(ctrl), zap.NewNop())
	err := w.HandleEvent(context.Background(),
		domain.EventEnvelope{Type: domain.EventTypeAvatarUploaded, Payload: json.RawMessage(`{broken`)})
	require.Error(t, err)
	require.True(t, broker.IsPermanent(err))
}

func TestHandleEvent_UnknownTypeIsIgnored(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := New(mocks.NewMockGophprofileRepository(ctrl), mocks.NewMockFileStorage(ctrl), zap.NewNop())
	err := w.HandleEvent(context.Background(),
		domain.EventEnvelope{Type: "avatar.future_thing", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
}
