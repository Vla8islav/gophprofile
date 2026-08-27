package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (s *PostgresStorage) ListAvatarsByUserID(ctx context.Context, userID int64) ([]domain.Avatar, error) {
	var avatars []domain.Avatar

	err := s.withRetry(ctx, func() error {
		rows, err := s.db.QueryContext(ctx,
			`SELECT `+avatarColumns+`
			 FROM avatars
			 WHERE user_id = $1 AND deleted_at IS NULL
			 ORDER BY created_at DESC`,
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to list avatars for user %d: %w", userID, err)
		}
		defer rows.Close()

		avatars = avatars[:0] // reset on retry so attempts don't accumulate
		for rows.Next() {
			var avatar domain.Avatar
			var thumbnails []byte
			if err := rows.Scan(
				&avatar.ID,
				&avatar.UserID,
				&avatar.FileName,
				&avatar.MimeType,
				&avatar.SizeBytes,
				&avatar.S3Key,
				&thumbnails,
				&avatar.UploadStatus,
				&avatar.ProcessingStatus,
				&avatar.CreatedAt,
				&avatar.UpdatedAt,
			); err != nil {
				return fmt.Errorf("failed to scan avatar row for user %d: %w", userID, err)
			}
			if len(thumbnails) > 0 {
				if err := json.Unmarshal(thumbnails, &avatar.ThumbnailS3Keys); err != nil {
					return fmt.Errorf("failed to decode thumbnail_s3_keys for avatar %s: %w", avatar.ID, err)
				}
			}
			avatars = append(avatars, avatar)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return avatars, nil
}
