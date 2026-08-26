package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUploadAvatar_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	var capturedID string
	repo.EXPECT().
		CreateAvatar(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params domain.CreateAvatarParams) (*domain.Avatar, error) {
			require.NoError(t, uuid.Validate(params.ID))
			require.Equal(t, int64(42), params.UserID)
			require.Equal(t, "cat.png", params.FileName)
			require.Equal(t, "image/png", params.MimeType)
			require.Equal(t, int64(4), params.SizeBytes)
			require.Equal(t, "avatars/"+params.ID+"/original", params.S3Key)
			capturedID = params.ID
			return &domain.Avatar{
				ID:               params.ID,
				UserID:           params.UserID,
				UploadStatus:     domain.UploadStatusUploading,
				ProcessingStatus: domain.ProcessingStatusPending,
			}, nil
		})
	storage.EXPECT().
		Upload(gomock.Any(), gomock.Any(), "image/png", int64(4), gomock.Any()).
		Return(nil)
	repo.EXPECT().
		SetAvatarUploadStatus(gomock.Any(), gomock.Any(), domain.UploadStatusCompleted).
		Return(nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	avatar, err := service.UploadAvatar(context.Background(), 42, "cat.png", "image/png", 4, bytes.NewReader([]byte("data")))
	require.NoError(t, err)
	require.Equal(t, capturedID, avatar.ID)
	require.Equal(t, domain.UploadStatusCompleted, avatar.UploadStatus)
	require.Equal(t, domain.ProcessingStatusPending, avatar.ProcessingStatus)
}

func TestUploadAvatar_RejectsUnsupportedMimeType(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := gophprofileService{
		repository:  mocks.NewMockGophprofileRepository(ctrl),
		fileStorage: mocks.NewMockFileStorage(ctrl),
	}

	_, err := service.UploadAvatar(context.Background(), 42, "evil.gif", "image/gif", 4, bytes.NewReader([]byte("data")))
	require.ErrorIs(t, err, domain.ErrUnsupportedAvatarFormat)
}

func TestUploadAvatar_StorageFailureSoftDeletesRecord(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	repo.EXPECT().
		CreateAvatar(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params domain.CreateAvatarParams) (*domain.Avatar, error) {
			return &domain.Avatar{ID: params.ID}, nil
		})
	storage.EXPECT().
		Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("minio down"))
	repo.EXPECT().
		SoftDeleteAvatar(gomock.Any(), gomock.Any()).
		Return(nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	_, err := service.UploadAvatar(context.Background(), 42, "cat.png", "image/png", 4, bytes.NewReader([]byte("data")))
	require.Error(t, err)
	require.Contains(t, err.Error(), "minio down")
}
