package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
)

func TestgophprofileService_UpdateSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repository}

	const userID = int64(42)
	id := uuid.New()
	req := domain.UpdateSecretRequest{
		Payload: []byte("v2"),
		Meta:    []byte("m2"),
		Version: 1,
	}

	repository.EXPECT().
		UpdateSecret(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params domain.UpdateSecretParams) (int64, error) {
			// Trust boundary: userID from the token, id from the URL - both land correctly,
			// and the body fields map through.
			require.Equal(t, userID, params.UserID)
			require.Equal(t, id, params.ID)
			require.Equal(t, req.Payload, params.Payload)
			require.Equal(t, req.Meta, params.Meta)
			require.Equal(t, req.Version, params.Version)
			return 2, nil
		})

	newVersion, err := svc.UpdateSecret(context.Background(), userID, id, req)
	require.NoError(t, err)
	require.Equal(t, int64(2), newVersion)
}

func TestgophprofileService_UpdateSecret_NilID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockgophprofileRepository(ctrl)
	svc := gophprofileService{repository: repository}

	// No EXPECT - repo must not be reached when the id is invalid.
	newVersion, err := svc.UpdateSecret(context.Background(), 42, uuid.Nil, domain.UpdateSecretRequest{Version: 1})
	require.ErrorIs(t, err, domain.ErrInvalidSecretID)
	require.Zero(t, newVersion)
}

func TestgophprofileService_UpdateSecret_PassesThroughErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
	}{
		{"version conflict", domain.ErrVersionConflict},
		{"not found", domain.ErrSecretNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repository := mocks.NewMockgophprofileRepository(ctrl)
			svc := gophprofileService{repository: repository}

			id := uuid.New()
			repository.EXPECT().
				UpdateSecret(gomock.Any(), gomock.Any()).
				Return(int64(0), tt.repoErr)

			newVersion, err := svc.UpdateSecret(context.Background(), 42, id, domain.UpdateSecretRequest{Version: 1})
			require.ErrorIs(t, err, tt.repoErr) // through the %w wrap
			require.Zero(t, newVersion)
		})
	}
}
