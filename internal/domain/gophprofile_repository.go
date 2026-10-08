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
	// ProcessUnsentOutboxEvents feeds unsent events oldest-first to handle,
	// marking each sent on success. handle runs inside the storage transaction as a
	// deliberate trade-off: this gives strict publish order across relay
	// instances, at the cost of holding a storage connection while a handle runs,
	// so callers must bound the batch via ctx. Delivery is at-least-once: an
	// aborted batch is re-delivered next cycle - handlers must stay idempotent.
	ProcessUnsentOutboxEvents(ctx context.Context, limit int,
		handle func(context.Context, OutboxEvent) error) (processed, fetched int, err error)
	SetAvatarProcessingStatus(ctx context.Context, avatarID string, status string) error
	SetAvatarThumbnails(ctx context.Context, avatarID string, thumbnailKeys map[string]string) error
	SoftDeleteAvatar(ctx context.Context, avatarID string) error
}
