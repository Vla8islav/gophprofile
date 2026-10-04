package domain

import (
	"context"
)

type GophprofileRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user CreateUserParams) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)

	CreateAvatar(ctx context.Context, params CreateAvatarParams) (*Avatar, error)
	GetAvatarByID(ctx context.Context, avatarID string) (*Avatar, error)
	GetLatestAvatarByUserID(ctx context.Context, userID int64) (*Avatar, error)
	ListAvatarsByUserID(ctx context.Context, userID int64) ([]Avatar, error)
	SetAvatarUploadStatus(ctx context.Context, avatarID string, status string) error
	CompleteAvatarUpload(ctx context.Context, avatarID string, event OutboxEvent) error
	SoftDeleteAvatarWithEvent(ctx context.Context, avatarID string, event OutboxEvent) error
	ProcessUnsentOutboxEvents(ctx context.Context, limit int,
		handle func(context.Context, OutboxEvent) error) (int, error)
	SetAvatarProcessingStatus(ctx context.Context, avatarID string, status string) error
	SetAvatarThumbnails(ctx context.Context, avatarID string, thumbnailKeys map[string]string) error
	SoftDeleteAvatar(ctx context.Context, avatarID string) error
}
