package repository

import (
	"context"
	"encoding/json"
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

func (s *PostgresStorage) SetAvatarProcessingStatus(ctx context.Context, avatarID string, status string) error {
	return s.withRetry(ctx, func() error {
		result, err := s.db.ExecContext(ctx,
			`UPDATE avatars
			 SET processing_status = $2, updated_at = now()
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID, status,
		)
		if err != nil {
			return fmt.Errorf("failed to set processing status for avatar %s: %w", avatarID, err)
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

// SetAvatarThumbnails records the generated thumbnail keys and marks completion
func (s *PostgresStorage) SetAvatarThumbnails(ctx context.Context, avatarID string, thumbnailKeys map[string]string) error {
	keysJSON, err := json.Marshal(thumbnailKeys)
	if err != nil {
		return fmt.Errorf("failed to encode thumbnail keys for avatar %s: %w", avatarID, err)
	}

	return s.withRetry(ctx, func() error {
		result, err := s.db.ExecContext(ctx,
			`UPDATE avatars
			 SET thumbnail_s3_keys = $2,
			     processing_status = $3,
			     updated_at = now()
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID, keysJSON, domain.ProcessingStatusCompleted,
		)
		if err != nil {
			return fmt.Errorf("failed to set thumbnails for avatar %s: %w", avatarID, err)
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
