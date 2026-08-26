package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// DeleteAvatar soft-deletes the DB record after an ownership check. Worker does an actual thing
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
	return nil
}
