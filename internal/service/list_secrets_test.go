package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
)

func TestgophprofileService_ListSecrets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repository}

	const userID = int64(42)
	want := []domain.SecretSummary{
		{ID: uuid.New(), Type: domain.SecretTypeText, Version: 1},
		{ID: uuid.New(), Type: domain.SecretTypeCard, Version: 1},
	}

	repository.EXPECT().
		ListSecrets(gomock.Any(), userID).
		Return(want, nil)

	got, err := svc.ListSecrets(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestgophprofileService_ListSecrets_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repository}

	repository.EXPECT().
		ListSecrets(gomock.Any(), int64(42)).
		Return(nil, nil)

	got, err := svc.ListSecrets(context.Background(), 42)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestgophprofileService_ListSecrets_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repository}

	repoErr := errors.New("db down")
	repository.EXPECT().
		ListSecrets(gomock.Any(), int64(42)).
		Return(nil, repoErr)

	got, err := svc.ListSecrets(context.Background(), 42)
	require.ErrorIs(t, err, repoErr) // matches through the %w wrap
	require.Nil(t, got)
}
