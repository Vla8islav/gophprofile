package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetricsService_LoginUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	passwordHash, err := helpers.HashPassword("test-password")
	require.NoError(t, err)

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		GetUserByLogin(gomock.Any(), "test-login").
		Return(&domain.User{
			ID:           123,
			Login:        "test-login",
			PasswordHash: passwordHash,
		}, nil)

	service := gophprofileService{
		repository: repository,
	}

	authResult, err := service.LoginUser(ctx, domain.UserLoginRequest{
		Login:    "test-login",
		Password: "test-password",
	})
	require.NoError(t, err)
	require.NotNil(t, authResult)
	require.Equal(t, int64(123), authResult.UserID)
	require.NotEmpty(t, authResult.Token)
}

func TestMetricsService_LoginUser_InvalidPassword(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	passwordHash, err := helpers.HashPassword("test-password")
	require.NoError(t, err)

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		GetUserByLogin(gomock.Any(), "test-login").
		Return(&domain.User{
			ID:           123,
			Login:        "test-login",
			PasswordHash: passwordHash,
		}, nil)

	service := gophprofileService{
		repository: repository,
	}

	authResult, err := service.LoginUser(ctx, domain.UserLoginRequest{
		Login:    "test-login",
		Password: "wrong-password",
	})
	require.ErrorIs(t, err, domain.ErrInvalidUserCredentials)
	require.Nil(t, authResult)
}

func TestMetricsService_LoginUser_UserNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		GetUserByLogin(gomock.Any(), "missing-login").
		Return(nil, domain.ErrInvalidUserCredentials)

	service := gophprofileService{
		repository: repository,
	}

	authResult, err := service.LoginUser(ctx, domain.UserLoginRequest{
		Login:    "missing-login",
		Password: "test-password",
	})
	require.ErrorIs(t, err, domain.ErrInvalidUserCredentials)
	require.Nil(t, authResult)
}

func TestMetricsService_LoginUser_RepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	repositoryErr := errors.New("repository error")

	repository := mocks.NewMockGophprofileRepository(ctrl)
	repository.EXPECT().
		GetUserByLogin(gomock.Any(), "test-login").
		Return(nil, repositoryErr)

	service := gophprofileService{
		repository: repository,
	}

	authResult, err := service.LoginUser(ctx, domain.UserLoginRequest{
		Login:    "test-login",
		Password: "test-password",
	})
	require.ErrorIs(t, err, repositoryErr)
	require.Nil(t, authResult)
}
