// Package worker is the consume side of the avatar pipeline: generates thumbnails and cleans up S3
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.uber.org/zap"
)

const (
	maxAttempts   = 3
	retryBaseWait = time.Second
)

type Worker struct {
	repository  domain.GophprofileRepository
	fileStorage domain.FileStorage
	logger      *zap.Logger
}

func New(repository domain.GophprofileRepository, fileStorage domain.FileStorage, logger *zap.Logger) *Worker {
	return &Worker{repository: repository, fileStorage: fileStorage, logger: logger}
}

// HandleEvent dispatches one envelope
func (w *Worker) HandleEvent(ctx context.Context, envelope domain.EventEnvelope) error {
	switch envelope.Type {
	case domain.EventTypeAvatarUploaded:
		var event domain.AvatarUploadEvent
		if err := json.Unmarshal(envelope.Payload, &event); err != nil {
			return fmt.Errorf("decode %s payload: %w", envelope.Type, err)
		}
		return w.handleUploaded(ctx, event)
	case domain.EventTypeAvatarDeleted:
		var event domain.AvatarDeleteEvent
		if err := json.Unmarshal(envelope.Payload, &event); err != nil {
			return fmt.Errorf("decode %s payload: %w", envelope.Type, err)
		}
		return w.handleDeleted(ctx, event)
	default:
		// Unknown types are logged
		w.logger.Warn("ignoring unknown event type", zap.String("type", envelope.Type))
		return nil
	}
}

// handleUploaded generates thumbnails
func (w *Worker) handleUploaded(ctx context.Context, event domain.AvatarUploadEvent) error {
	avatar, err := w.repository.GetAvatarByID(ctx, event.AvatarID)
	if errors.Is(err, domain.ErrAvatarNotFound) {
		w.logger.Info("avatar gone before processing, skipping",
			zap.String("avatar_id", event.AvatarID))
		return nil
	}
	if err != nil {
		return fmt.Errorf("load avatar %s: %w", event.AvatarID, err)
	}

	// Idempotency: a redelivered event for an already-processed avatar is a no-op.
	if avatar.ProcessingStatus == domain.ProcessingStatusCompleted {
		w.logger.Info("avatar already processed, skipping",
			zap.String("avatar_id", event.AvatarID))
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = w.generateThumbnails(ctx, avatar)
		if lastErr == nil {
			w.logger.Info("thumbnails generated",
				zap.String("avatar_id", event.AvatarID))
			return nil
		}
		w.logger.Warn("thumbnail generation attempt failed",
			zap.String("avatar_id", event.AvatarID),
			zap.Int("attempt", attempt),
			zap.Error(lastErr),
		)
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * retryBaseWait):
			}
		}
	}

	if err := w.repository.SetAvatarProcessingStatus(ctx, event.AvatarID, domain.ProcessingStatusFailed); err != nil {
		w.logger.Error("failed to mark avatar as failed",
			zap.String("avatar_id", event.AvatarID), zap.Error(err))
	}
	return fmt.Errorf("thumbnails for avatar %s after %d attempts: %w", event.AvatarID, maxAttempts, lastErr)
}

// handleDeleted removes every S3 object the avatar owned
func (w *Worker) handleDeleted(ctx context.Context, event domain.AvatarDeleteEvent) error {
	var failed []string
	for _, key := range event.S3Keys {
		if err := w.fileStorage.Delete(ctx, key); err != nil {
			w.logger.Warn("failed to delete object",
				zap.String("avatar_id", event.AvatarID),
				zap.String("key", key),
				zap.Error(err),
			)
			failed = append(failed, key)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("delete %d/%d objects for avatar %s", len(failed), len(event.S3Keys), event.AvatarID)
	}
	w.logger.Info("avatar objects deleted",
		zap.String("avatar_id", event.AvatarID),
		zap.Int("count", len(event.S3Keys)))
	return nil
}
