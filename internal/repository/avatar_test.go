package repository

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func createTestUser(t *testing.T, storage *PostgresStorage, prefix string) int64 {
	t.Helper()
	userID, err := storage.CreateUser(context.Background(), domain.CreateUserParams{
		Login:        helpers.UniqueLogin(prefix),
		PasswordHash: "hashed-password",
	})
	require.NoError(t, err)
	return userID
}

func createTestAvatar(t *testing.T, storage *PostgresStorage, userID int64) *domain.Avatar {
	t.Helper()
	avatarID := uuid.NewString()
	avatar, err := storage.CreateAvatar(context.Background(), domain.CreateAvatarParams{
		ID:        avatarID,
		UserID:    userID,
		FileName:  "cat.png",
		MimeType:  "image/png",
		SizeBytes: 1024,
		S3Key:     "avatars/" + avatarID + "/original",
	})
	require.NoError(t, err)
	return avatar
}

func TestPostgresStorage_CreateAndGetAvatar(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, storage, "avatar-create-test")
	created := createTestAvatar(t, storage, userID)

	require.Equal(t, domain.UploadStatusUploading, created.UploadStatus)
	require.Equal(t, domain.ProcessingStatusPending, created.ProcessingStatus)
	require.False(t, created.CreatedAt.IsZero())

	fetched, err := storage.GetAvatarByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, userID, fetched.UserID)
	require.Equal(t, "cat.png", fetched.FileName)
	require.Equal(t, "image/png", fetched.MimeType)
	require.Equal(t, int64(1024), fetched.SizeBytes)
	require.Nil(t, fetched.ThumbnailS3Keys)
}

func TestPostgresStorage_GetAvatarByID_NotFound(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	_, err := storage.GetAvatarByID(context.Background(), uuid.NewString())
	require.ErrorIs(t, err, domain.ErrAvatarNotFound)
}

func TestPostgresStorage_SetAvatarUploadStatus(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, storage, "avatar-status-test")
	avatar := createTestAvatar(t, storage, userID)

	require.NoError(t, storage.SetAvatarUploadStatus(ctx, avatar.ID, domain.UploadStatusCompleted))

	fetched, err := storage.GetAvatarByID(ctx, avatar.ID)
	require.NoError(t, err)
	require.Equal(t, domain.UploadStatusCompleted, fetched.UploadStatus)
	require.True(t, fetched.UpdatedAt.After(avatar.UpdatedAt) || fetched.UpdatedAt.Equal(avatar.UpdatedAt))
}

func TestPostgresStorage_SoftDeleteAvatar(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, storage, "avatar-delete-test")
	avatar := createTestAvatar(t, storage, userID)

	require.NoError(t, storage.SoftDeleteAvatar(ctx, avatar.ID))

	// Soft-deleted avatars are invisible to reads...
	_, err := storage.GetAvatarByID(ctx, avatar.ID)
	require.ErrorIs(t, err, domain.ErrAvatarNotFound)

	// ...but the row itself survives (deleted_at is set, not a DELETE).
	var deletedAtSet bool
	err = storage.db.QueryRowContext(ctx,
		`SELECT deleted_at IS NOT NULL FROM avatars WHERE id = $1`,
		avatar.ID,
	).Scan(&deletedAtSet)
	require.NoError(t, err)
	require.True(t, deletedAtSet)

	// Double delete reports not-found.
	require.ErrorIs(t, storage.SoftDeleteAvatar(ctx, avatar.ID), domain.ErrAvatarNotFound)
}

func TestPostgresStorage_GetLatestAvatarByUserID(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, storage, "avatar-latest-test")

	_, err := storage.GetLatestAvatarByUserID(ctx, userID)
	require.ErrorIs(t, err, domain.ErrAvatarNotFound)

	first := createTestAvatar(t, storage, userID)
	second := createTestAvatar(t, storage, userID)

	// Force distinct created_at ordering: within one transaction-less burst
	// now() can tie, so nudge the second row explicitly.
	_, err = storage.db.ExecContext(ctx,
		`UPDATE avatars SET created_at = created_at + interval '1 second' WHERE id = $1`,
		second.ID)
	require.NoError(t, err)

	latest, err := storage.GetLatestAvatarByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, second.ID, latest.ID)

	// Deleting the latest must surface the previous one.
	require.NoError(t, storage.SoftDeleteAvatar(ctx, second.ID))
	latest, err = storage.GetLatestAvatarByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, first.ID, latest.ID)
}

func TestPostgresStorage_ListAvatarsByUserID(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, storage, "avatar-list-test")

	avatars, err := storage.ListAvatarsByUserID(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, avatars)

	first := createTestAvatar(t, storage, userID)
	second := createTestAvatar(t, storage, userID)

	avatars, err = storage.ListAvatarsByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, avatars, 2)

	// Soft-deleted ones drop out of the list.
	require.NoError(t, storage.SoftDeleteAvatar(ctx, first.ID))
	avatars, err = storage.ListAvatarsByUserID(ctx, userID)
	require.NoError(t, err)
	require.Len(t, avatars, 1)
	require.Equal(t, second.ID, avatars[0].ID)
}

func TestPostgresStorage_CreateAvatar_UnknownUserFails(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	avatarID := uuid.NewString()
	_, err := storage.CreateAvatar(context.Background(), domain.CreateAvatarParams{
		ID:        avatarID,
		UserID:    999999999, // no such user: FK must reject
		FileName:  "cat.png",
		MimeType:  "image/png",
		SizeBytes: 1024,
		S3Key:     "avatars/" + avatarID + "/original",
	})
	require.Error(t, err)
}
