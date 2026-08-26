package service

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestDeleteAvatar_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "id-1").
		Return(&domain.Avatar{ID: "id-1", UserID: 42}, nil)
	repo.EXPECT().SoftDeleteAvatar(gomock.Any(), "id-1").Return(nil)
	events := mocks.NewMockEventPublisher(ctrl)
	events.EXPECT().
		Publish(gomock.Any(), "id-1", domain.EventTypeAvatarDeleted, gomock.Any()).
		Return(nil)

	service := gophprofileService{repository: repo, events: events, logger: zap.NewNop()}

	require.NoError(t, service.DeleteAvatar(context.Background(), "id-1", 42))
}

func TestDeleteAvatar_ForbiddenForOtherUser(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "id-1").
		Return(&domain.Avatar{ID: "id-1", UserID: 42}, nil)

	service := gophprofileService{repository: repo}

	err := service.DeleteAvatar(context.Background(), "id-1", 43)
	require.ErrorIs(t, err, domain.ErrNotAvatarOwner)
}

func TestDeleteAvatar_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetAvatarByID(gomock.Any(), "missing").
		Return(nil, domain.ErrAvatarNotFound)

	service := gophprofileService{repository: repo}

	err := service.DeleteAvatar(context.Background(), "missing", 42)
	require.ErrorIs(t, err, domain.ErrAvatarNotFound)
}

func TestDeleteUserAvatar_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockGophprofileRepository(ctrl)
	repo.EXPECT().GetLatestAvatarByUserID(gomock.Any(), int64(42)).
		Return(&domain.Avatar{ID: "id-1", UserID: 42}, nil)
	repo.EXPECT().SoftDeleteAvatar(gomock.Any(), "id-1").Return(nil)
	events := mocks.NewMockEventPublisher(ctrl)
	events.EXPECT().
		Publish(gomock.Any(), "id-1", domain.EventTypeAvatarDeleted, gomock.Any()).
		Return(nil)

	service := gophprofileService{repository: repo, events: events, logger: zap.NewNop()}

	require.NoError(t, service.DeleteUserAvatar(context.Background(), 42, 42))
}

func TestDeleteUserAvatar_ForbiddenForOtherUser(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := gophprofileService{repository: mocks.NewMockGophprofileRepository(ctrl)}

	err := service.DeleteUserAvatar(context.Background(), 42, 43)
	require.ErrorIs(t, err, domain.ErrNotAvatarOwner)
}
