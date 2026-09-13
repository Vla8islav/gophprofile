package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func enqueueOutboxTx(ctx context.Context, tx *sql.Tx, event domain.OutboxEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO outbox_events (event_key, event_type, payload)
		 VALUES ($1, $2, $3)`,
		event.Key, event.Type, []byte(event.Payload),
	)
	if err != nil {
		return fmt.Errorf("enqueue outbox event %s: %w", event.Type, err)
	}
	return nil
}

// CompleteAvatarUpload marks the upload finished and enqueues the avatar.uploaded event in ONE transaction
func (s *PostgresStorage) CompleteAvatarUpload(ctx context.Context, avatarID string, event domain.OutboxEvent) error {
	return s.withRetryTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`UPDATE avatars
			 SET upload_status = $2, updated_at = now()
			 WHERE id = $1 AND deleted_at IS NULL`,
			avatarID, domain.UploadStatusCompleted,
		)
		if err != nil {
			return fmt.Errorf("failed to complete upload for avatar %s: %w", avatarID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows for avatar %s: %w", avatarID, err)
		}
		if affected == 0 {
			return domain.ErrAvatarNotFound
		}
		return enqueueOutboxTx(ctx, tx, event)
	})
}

// SoftDeleteAvatarWithEvent hides the avatar and enqueues the avatar.deleted event
func (s *PostgresStorage) SoftDeleteAvatarWithEvent(ctx context.Context, avatarID string, event domain.OutboxEvent) error {
	return s.withRetryTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
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
		return enqueueOutboxTx(ctx, tx, event)
	})
}

// UnsentOutboxEvents returns the oldest unpublished events in insertion
func (s *PostgresStorage) UnsentOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent

	err := s.withRetry(ctx, func() error {
		rows, err := s.db.QueryContext(ctx,
			`SELECT id, event_key, event_type, payload, created_at
			 FROM outbox_events
			 WHERE sent_at IS NULL
			 ORDER BY id
			 LIMIT $1`,
			limit,
		)
		if err != nil {
			return fmt.Errorf("failed to list unsent outbox events: %w", err)
		}
		defer rows.Close()

		events = events[:0]
		for rows.Next() {
			var event domain.OutboxEvent
			var payload []byte
			if err := rows.Scan(&event.ID, &event.Key, &event.Type, &payload, &event.CreatedAt); err != nil {
				return fmt.Errorf("failed to scan outbox event: %w", err)
			}
			event.Payload = payload
			events = append(events, event)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *PostgresStorage) MarkOutboxEventSent(ctx context.Context, eventID int64) error {
	return s.withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			`UPDATE outbox_events SET sent_at = now() WHERE id = $1`,
			eventID,
		)
		if err != nil {
			return fmt.Errorf("failed to mark outbox event %d sent: %w", eventID, err)
		}
		return nil
	})
}
