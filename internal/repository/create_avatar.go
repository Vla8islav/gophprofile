package repository

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (s *PostgresStorage) CreateAvatar(ctx context.Context, params domain.CreateAvatarParams) (*domain.Avatar, error) {
	avatar := domain.Avatar{
		ID:        params.ID,
		UserID:    params.UserID,
		FileName:  params.FileName,
		MimeType:  params.MimeType,
		SizeBytes: params.SizeBytes,
		S3Key:     params.S3Key,
	}

	err := s.withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx,
			`INSERT INTO avatars (id, user_id, file_name, mime_type, size_bytes, s3_key)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING upload_status, processing_status, created_at, updated_at`,
			params.ID,
			params.UserID,
			params.FileName,
			params.MimeType,
			params.SizeBytes,
			params.S3Key,
		).Scan(&avatar.UploadStatus, &avatar.ProcessingStatus, &avatar.CreatedAt, &avatar.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create avatar for user %d: %w", params.UserID, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &avatar, nil
}
