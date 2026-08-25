package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	maxRetryAttempts = 3
	baseRetryDelay   = 50 * time.Millisecond
)

// backoff provides exponential backoff with a base delay of 50ms.
func backoff(ctx context.Context, attempt int) error {
	delay := baseRetryDelay << (attempt - 1)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

func (s *PostgresStorage) withRetry(ctx context.Context, executeFunction func() error) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := executeFunction()
		if err == nil {
			return nil
		}

		if s.isRetriablePostgresError(err) && attempt < maxRetryAttempts {
			lastErr = err
			if waitErr := backoff(ctx, attempt); waitErr != nil {
				return waitErr // context cancelled during the wait
			}
			continue
		}

		return err
	}

	return fmt.Errorf("operation failed after retries: %w", lastErr)
}

func (s *PostgresStorage) withRetryTx(ctx context.Context, executeFunction func(*sql.Tx) error) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			if s.isRetriablePostgresError(err) && attempt < maxRetryAttempts {
				lastErr = err
				continue
			}
			return fmt.Errorf("begin tx: %w", err)
		}

		err = executeFunction(tx)
		if err != nil {
			_ = tx.Rollback()

			if s.isRetriablePostgresError(err) && attempt < maxRetryAttempts {
				lastErr = err
				continue
			}

			return err
		}

		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()

			if s.isRetriablePostgresError(err) && attempt < maxRetryAttempts {
				lastErr = err
				continue
			}

			return fmt.Errorf("commit tx: %w", err)
		}

		return nil
	}

	return fmt.Errorf("transaction failed after retries: %w", lastErr)
}
