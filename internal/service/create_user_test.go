package service

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGophprofileService_CreateUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	userRegReq := domain.UserRegisterRequest{
		Login:    "test-login",
		Password: "test-password",
	}

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params domain.CreateUserParams) (int64, error) {
			require.Equal(t, userRegReq.Login, params.Login)
			require.NotEqual(t, userRegReq.Password, params.PasswordHash)
			require.NoError(t, helpers.CompareHashAndPassword(params.PasswordHash, userRegReq.Password))

			return 123, nil
		})

	service := gophprofileService{
		repository: repository,
		authSecret: []byte("test-secret"),
	}

	authResult, err := service.CreateUser(ctx, userRegReq)
	require.NoError(t, err)
	require.NotNil(t, authResult)
	require.Equal(t, int64(123), authResult.UserID)
	require.NotEmpty(t, authResult.Token)
}

func TestGophprofileService_CreateUser_HashesSamePasswordDifferently(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	firstReq := domain.UserRegisterRequest{Login: "first-user", Password: "same-password"}
	secondReq := domain.UserRegisterRequest{Login: "second-user", Password: "same-password"}

	var passwordHashes []string

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params domain.CreateUserParams) (int64, error) {
			passwordHashes = append(passwordHashes, params.PasswordHash)
			return int64(len(passwordHashes)), nil
		}).
		Times(2)

	service := gophprofileService{
		repository: repository,
		authSecret: []byte("test-secret"),
	}

	firstAuthResult, err := service.CreateUser(ctx, firstReq)
	require.NoError(t, err)
	require.NotNil(t, firstAuthResult)
	require.Equal(t, int64(1), firstAuthResult.UserID)
	require.NotEmpty(t, firstAuthResult.Token)

	secondAuthResult, err := service.CreateUser(ctx, secondReq)
	require.NoError(t, err)
	require.NotNil(t, secondAuthResult)
	require.Equal(t, int64(2), secondAuthResult.UserID)
	require.NotEmpty(t, secondAuthResult.Token)

	require.Len(t, passwordHashes, 2)
	require.NotEqual(t, firstReq.Password, passwordHashes[0])
	require.NotEqual(t, secondReq.Password, passwordHashes[1])
	require.NotEqual(t, passwordHashes[0], passwordHashes[1])
	require.NoError(t, helpers.CompareHashAndPassword(passwordHashes[0], firstReq.Password))
	require.NoError(t, helpers.CompareHashAndPassword(passwordHashes[1], secondReq.Password))
}
