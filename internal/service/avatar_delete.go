package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// deleteEventFor lists every S3 object the avatar owns
func deleteEventFor(avatar *domain.Avatar) domain.AvatarDeleteEvent {
	keys := make([]string, 0, 1+len(avatar.ThumbnailS3Keys))
	keys = append(keys, avatar.S3Key)
	for _, key := range avatar.ThumbnailS3Keys {
		keys = append(keys, key)
	}
	return domain.AvatarDeleteEvent{AvatarID: avatar.ID, S3Keys: keys}
}

// DeleteAvatar soft-deletes the DB record after an ownership check
// S3 objects are being removed asynchronously by the worker.
func (m gophprofileService) DeleteAvatar(ctx context.Context, avatarID string, requesterID int64) error {
	avatar, err := m.repository.GetAvatarByID(ctx, avatarID)
	if err != nil {
		return err
	}
	if avatar.UserID != requesterID {
		return domain.ErrNotAvatarOwner
	}

	if err := m.repository.SoftDeleteAvatar(ctx, avatarID); err != nil {
		return fmt.Errorf("failed to delete avatar %s: %w", avatarID, err)
	}

	m.publishEvent(ctx, avatar.ID, domain.EventTypeAvatarDeleted, deleteEventFor(avatar))
	return nil
}

func (m gophprofileService) DeleteUserAvatar(ctx context.Context, userID int64, requesterID int64) error {
	if userID != requesterID {
		return domain.ErrNotAvatarOwner
	}

	avatar, err := m.repository.GetLatestAvatarByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if err := m.repository.SoftDeleteAvatar(ctx, avatar.ID); err != nil {
		return fmt.Errorf("failed to delete avatar %s for user %d: %w", avatar.ID, userID, err)
	}

	m.publishEvent(ctx, avatar.ID, domain.EventTypeAvatarDeleted, deleteEventFor(avatar))
	return nil
}
