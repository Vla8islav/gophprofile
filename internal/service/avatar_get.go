package service

import (
	"context"
	"fmt"
	"io"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// contentKeyFor picks the S3 key for the requested size variant
func contentKeyFor(avatar *domain.Avatar, sizeVariant string) string {
	// fallback:
	if sizeVariant == "" || sizeVariant == "original" {
		return avatar.S3Key
	}
	if key, ok := avatar.ThumbnailS3Keys[sizeVariant]; ok {
		return key
	}
	return avatar.S3Key
}

func (m gophprofileService) GetAvatarContent(ctx context.Context, avatarID string, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	avatar, err := m.repository.GetAvatarByID(ctx, avatarID)
	if err != nil {
		return nil, nil, err
	}

	content, err := m.fileStorage.Download(ctx, contentKeyFor(avatar, sizeVariant))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download avatar %s: %w", avatarID, err)
	}

	return avatar, content, nil
}

func (m gophprofileService) GetUserAvatarContent(ctx context.Context, userID int64, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	avatar, err := m.repository.GetLatestAvatarByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	content, err := m.fileStorage.Download(ctx, contentKeyFor(avatar, sizeVariant))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download avatar for user %d: %w", userID, err)
	}

	return avatar, content, nil
}

func (m gophprofileService) GetAvatarMetadata(ctx context.Context, avatarID string) (*domain.Avatar, error) {
	return m.repository.GetAvatarByID(ctx, avatarID)
}

func (m gophprofileService) ListUserAvatars(ctx context.Context, userID int64) ([]domain.Avatar, error) {
	return m.repository.ListAvatarsByUserID(ctx, userID)
}
