package repository

import (
	"context"
	"fmt"
)

func (s *PostgresStorage) Ping(ctx context.Context) error {

	// verify connection
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("couldn't ping postgres db: %w", err)
	}

	return nil
}
