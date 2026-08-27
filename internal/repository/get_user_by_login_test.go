package repository

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPostgresStorage_GetUserByLogin(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	params := domain.CreateUserParams{
		Login:        helpers.UniqueLogin("get-user-by-login-test"),
		PasswordHash: "hashed-password",
	}

	userID, err := storage.CreateUser(ctx, params)
	require.NoError(t, err)
	require.Greater(t, userID, int64(0))

	user, err := storage.GetUserByLogin(ctx, params.Login)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.Equal(t, params.Login, user.Login)
	require.Equal(t, params.PasswordHash, user.PasswordHash)
}

func TestPostgresStorage_GetUserByLogin_NotFound(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	user, err := storage.GetUserByLogin(ctx, helpers.UniqueLogin("missing-user"))
	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, user)
}
