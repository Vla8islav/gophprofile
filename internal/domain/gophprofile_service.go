package domain

import (
	"context"
	"io"
)

type GophprofileService interface {
	Ping(ctx context.Context) error
	FileStoragePing(ctx context.Context) error
	BrokerPing(ctx context.Context) error
	CreateUser(ctx context.Context, request UserRegisterRequest) (*AuthResult, error)
	LoginUser(ctx context.Context, request UserLoginRequest) (*AuthResult, error)

	UploadAvatar(ctx context.Context, userID int64, fileName, mimeType string, size int64, content io.Reader) (*Avatar, error)
	GetAvatarContent(ctx context.Context, avatarID string, sizeVariant string) (*Avatar, io.ReadCloser, error)
	GetUserAvatarContent(ctx context.Context, userID int64, sizeVariant string) (*Avatar, io.ReadCloser, error)
	GetAvatarMetadata(ctx context.Context, avatarID string) (*Avatar, error)
	ListUserAvatars(ctx context.Context, userID int64) ([]Avatar, error)
	DeleteAvatar(ctx context.Context, avatarID string, requesterID int64) error
	DeleteUserAvatar(ctx context.Context, userID int64, requesterID int64) error
}
