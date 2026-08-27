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

func TestPostgresStorage_CreateUser(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	params := domain.CreateUserParams{
		Login:        helpers.UniqueLogin("create-user-test"),
		PasswordHash: "hashed-password",
	}

	userID, err := storage.CreateUser(ctx, params)
	require.NoError(t, err)
	require.Greater(t, userID, int64(0))

	var login string
	var passwordHash string

	err = storage.db.QueryRowContext(ctx,
		`SELECT login, password_hash
                 FROM users
                 WHERE id = $1`,
		userID,
	).Scan(&login, &passwordHash)
	require.NoError(t, err)

	require.Equal(t, params.Login, login)
	require.Equal(t, params.PasswordHash, passwordHash)
}

func TestPostgresStorage_CreateUser_DuplicateLogin(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	params := domain.CreateUserParams{
		Login:        helpers.UniqueLogin("duplicate-user-test"),
		PasswordHash: "hashed-password",
	}

	userID, err := storage.CreateUser(ctx, params)
	require.NoError(t, err)
	require.Greater(t, userID, int64(0))

	duplicateUserID, err := storage.CreateUser(ctx, params)
	require.Error(t, err)
	require.Zero(t, duplicateUserID)
}

func TestPostgresStorage_CreateUser_AllowsSamePasswordHash(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	firstParams := domain.CreateUserParams{
		Login:        helpers.UniqueLogin("same-password-user-1"),
		PasswordHash: "same-password-hash",
	}
	secondParams := domain.CreateUserParams{
		Login:        helpers.UniqueLogin("same-password-user-2"),
		PasswordHash: "same-password-hash",
	}

	firstUserID, err := storage.CreateUser(ctx, firstParams)
	require.NoError(t, err)
	require.Greater(t, firstUserID, int64(0))

	secondUserID, err := storage.CreateUser(ctx, secondParams)
	require.NoError(t, err)
	require.Greater(t, secondUserID, int64(0))
	require.NotEqual(t, firstUserID, secondUserID)
}
