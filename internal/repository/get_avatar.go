package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

const avatarColumns = `id, user_id, file_name, mime_type, size_bytes, s3_key,
	thumbnail_s3_keys, upload_status, processing_status, created_at, updated_at`

func scanAvatar(row *sql.Row) (*domain.Avatar, error) {
	var avatar domain.Avatar
	var thumbnails []byte

	err := row.Scan(
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
	)
	if err != nil {
		return nil, err
	}
	if len(thumbnails) > 0 {
		if err := json.Unmarshal(thumbnails, &avatar.ThumbnailS3Keys); err != nil {
			return nil, fmt.Errorf("failed to decode thumbnail_s3_keys for avatar %s: %w", avatar.ID, err)
		}
	}
	return &avatar, nil
}

func (s *PostgresStorage) GetAvatarByID(ctx context.Context, avatarID string) (*domain.Avatar, error) {
	var avatar *domain.Avatar

	err := s.withRetry(ctx, func() error {
		var err error
		avatar, err = scanAvatar(s.db.QueryRowContext(ctx,
			`SELECT `+avatarColumns+`
			 FROM avatars
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID,
		))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAvatarNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get avatar %s: %w", avatarID, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return avatar, nil
}

func (s *PostgresStorage) GetLatestAvatarByUserID(ctx context.Context, userID int64) (*domain.Avatar, error) {
	var avatar *domain.Avatar

	err := s.withRetry(ctx, func() error {
		var err error
		avatar, err = scanAvatar(s.db.QueryRowContext(ctx,
			`SELECT `+avatarColumns+`
			 FROM avatars
			 WHERE user_id = $1 AND deleted_at IS NULL
			 ORDER BY created_at DESC
			 LIMIT 1`,
			userID,
		))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAvatarNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get latest avatar for user %d: %w", userID, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return avatar, nil
}
