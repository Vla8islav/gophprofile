package service

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAvatarContent_ServesOriginal(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	stored := &domain.Avatar{ID: "id-1", S3Key: "avatars/id-1/original", MimeType: "image/png"}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "id-1").Return(stored, nil)
	storage.EXPECT().Download(gomock.Any(), "avatars/id-1/original").
		Return(io.NopCloser(strings.NewReader("png-bytes")), nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	avatar, content, err := service.GetAvatarContent(context.Background(), "id-1", "")
	require.NoError(t, err)
	defer content.Close()

	require.Equal(t, stored, avatar)
	data, err := io.ReadAll(content)
	require.NoError(t, err)
	require.Equal(t, "png-bytes", string(data))
}

func TestGetAvatarContent_ThumbnailFallsBackToOriginal(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	// No thumbnails generated yet: size=100x100 must serve the original.
	stored := &domain.Avatar{ID: "id-1", S3Key: "avatars/id-1/original"}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "id-1").Return(stored, nil)
	storage.EXPECT().Download(gomock.Any(), "avatars/id-1/original").
		Return(io.NopCloser(strings.NewReader("original")), nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	_, content, err := service.GetAvatarContent(context.Background(), "id-1", "100x100")
	require.NoError(t, err)
	_ = content.Close()
}

func TestGetAvatarContent_ServesThumbnailWhenPresent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	stored := &domain.Avatar{
		ID:              "id-1",
		S3Key:           "avatars/id-1/original",
		ThumbnailS3Keys: map[string]string{"100x100": "thumbnails/id-1/100x100.jpg"},
	}
	repo.EXPECT().GetAvatarByID(gomock.Any(), "id-1").Return(stored, nil)
	storage.EXPECT().Download(gomock.Any(), "thumbnails/id-1/100x100.jpg").
		Return(io.NopCloser(strings.NewReader("thumb")), nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	_, content, err := service.GetAvatarContent(context.Background(), "id-1", "100x100")
	require.NoError(t, err)
	_ = content.Close()
}

func TestGetAvatarContent_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "missing").Return(nil, domain.ErrAvatarNotFound)

	service := gophprofileService{repository: repo, fileStorage: mocks.NewMockFileStorage(ctrl)}

	_, _, err := service.GetAvatarContent(context.Background(), "missing", "")
	require.ErrorIs(t, err, domain.ErrAvatarNotFound)
}

func TestGetUserAvatarContent_UsesLatestAvatar(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	storage := mocks.NewMockFileStorage(ctrl)

	stored := &domain.Avatar{ID: "id-2", UserID: 42, S3Key: "avatars/id-2/original"}
	repo.EXPECT().GetLatestAvatarByUserID(gomock.Any(), int64(42)).Return(stored, nil)
	storage.EXPECT().Download(gomock.Any(), "avatars/id-2/original").
		Return(io.NopCloser(strings.NewReader("latest")), nil)

	service := gophprofileService{repository: repo, fileStorage: storage}

	avatar, content, err := service.GetUserAvatarContent(context.Background(), 42, "")
	require.NoError(t, err)
	_ = content.Close()
	require.Equal(t, "id-2", avatar.ID)
}
