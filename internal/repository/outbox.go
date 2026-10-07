package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func enqueueOutboxTx(ctx context.Context, tx *sql.Tx, event domain.OutboxEvent) error {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	var traceContext any
	if len(carrier) > 0 {
		b, err := json.Marshal(carrier)
		if err != nil {
			return fmt.Errorf("marshal trace context: %w", err)
		}
		traceContext = b
	}

	_, err := tx.ExecContext(ctx,
		`INSERT INTO outbox_events (event_key, event_type, payload, trace_context)
		 VALUES ($1, $2, $3, $4)`,
		event.Key, event.Type, []byte(event.Payload),
		traceContext,
	)
	if err != nil {
		return fmt.Errorf("enqueue outbox event %s: %w", event.Type, err)
	}
	return nil
}

// ProcessUnsentOutboxEvents implements the domain contract via one
// FOR UPDATE SKIP LOCKED transaction. On ctx timeout database rolls the
// tx back: sent_at marks are lost and already-published events repeat.
func (s *PostgresStorage) ProcessUnsentOutboxEvents(ctx context.Context, limit int,
	handle func(context.Context, domain.OutboxEvent) error) (processed, fetched int, err error) {

	// no withRetryTx: retrying the tx would re-run handle and republish to Kafka
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin outbox tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT id, event_key, event_type, payload, created_at, trace_context
                 FROM outbox_events
                 WHERE sent_at IS NULL
                 ORDER BY id
                 LIMIT $1
                 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("select unsent outbox events: %w", err)
	}

	events := []domain.OutboxEvent{}
	for rows.Next() {
		var event domain.OutboxEvent
		var payload []byte      // JSONB columns need a []byte
		var traceContext []byte // intermediary — can't scan into json.RawMessage
		if err := rows.Scan(
			&event.ID,
			&event.Key,
			&event.Type,
			&payload,
			&event.CreatedAt,
			&traceContext,
		); err != nil {
			rows.Close()
			return 0, 0, fmt.Errorf("scan outbox event: %w", err)
		}
		event.Payload = payload

		event.TraceContext = traceContext
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, len(events), fmt.Errorf("iterate outbox events: %w", err)
	}
	rows.Close()

	processed = 0
	for _, ev := range events {
		if err := handle(ctx, ev); err != nil {
			break // publish failed: stop, but still commit the ones that worked
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE outbox_events SET sent_at = now() WHERE id = $1`, ev.ID); err != nil {
			return processed, len(events), fmt.Errorf("mark outbox event %d sent: %w", ev.ID, err)
		}
		processed++
	}

	if err := tx.Commit(); err != nil {
		return 0, len(events), fmt.Errorf("commit outbox tx: %w", err)
	}
	return processed, len(events), nil
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
