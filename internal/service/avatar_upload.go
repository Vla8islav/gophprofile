package service

import (
	"context"
	"fmt"
	"io"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/google/uuid"
)

// UploadAvatar stores the original in the object store and records metadata.
// Thumbnail generation being done by the worker.
func (m gophprofileService) UploadAvatar(ctx context.Context, userID int64, fileName, mimeType string, size int64, content io.Reader) (*domain.Avatar, error) {
	if !domain.AllowedAvatarMimeTypes[mimeType] {
		return nil, fmt.Errorf("mime type %s: %w", mimeType, domain.ErrUnsupportedAvatarFormat)
	}

	avatarID := uuid.NewString()
	s3Key := fmt.Sprintf("avatars/%s/original", avatarID)

	avatar, err := m.repository.CreateAvatar(ctx, domain.CreateAvatarParams{
		ID:        avatarID,
		UserID:    userID,
		FileName:  fileName,
		MimeType:  mimeType,
		SizeBytes: size,
		S3Key:     s3Key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create avatar record for user %d: %w", userID, err)
	}

	if err := m.fileStorage.Upload(ctx, s3Key, mimeType, size, content); err != nil {
		// The DB row exists but the bytes never landed
		_ = m.repository.SoftDeleteAvatar(ctx, avatarID)
		return nil, fmt.Errorf("failed to upload avatar %s to file storage: %w", avatarID, err)
	}

	if err := m.repository.SetAvatarUploadStatus(ctx, avatarID, domain.UploadStatusCompleted); err != nil {
		return nil, fmt.Errorf("failed to mark avatar %s as uploaded: %w", avatarID, err)
	}
	avatar.UploadStatus = domain.UploadStatusCompleted

	m.publishEvent(ctx, avatarID, domain.EventTypeAvatarUploaded, domain.AvatarUploadEvent{
		AvatarID: avatarID,
		UserID:   userID,
		S3Key:    s3Key,
	})

	return avatar, nil
}
