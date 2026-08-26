package repository

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (s *PostgresStorage) SetAvatarUploadStatus(ctx context.Context, avatarID string, status string) error {
	return s.withRetry(ctx, func() error {
		result, err := s.db.ExecContext(ctx,
			`UPDATE avatars
			 SET upload_status = $2, updated_at = now()
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID, status,
		)
		if err != nil {
			return fmt.Errorf("failed to set upload status for avatar %s: %w", avatarID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows for avatar %s: %w", avatarID, err)
		}
		if affected == 0 {
			return domain.ErrAvatarNotFound
		}
		return nil
	})
}

func (s *PostgresStorage) SoftDeleteAvatar(ctx context.Context, avatarID string) error {
	return s.withRetry(ctx, func() error {
		result, err := s.db.ExecContext(ctx,
			`UPDATE avatars
			 SET deleted_at = now(), updated_at = now()
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID,
		)
		if err != nil {
			return fmt.Errorf("failed to soft-delete avatar %s: %w", avatarID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows for avatar %s: %w", avatarID, err)
		}
		if affected == 0 {
			return domain.ErrAvatarNotFound
		}
		return nil
	})
}
