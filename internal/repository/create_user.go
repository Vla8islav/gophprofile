package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (s *PostgresStorage) CreateUser(ctx context.Context, user domain.CreateUserParams) (int64, error) {
	var userID int64

	err := s.withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx,
			`INSERT INTO users (login, password_hash)
			 VALUES ($1, $2)
			 ON CONFLICT (login) DO NOTHING
			 RETURNING id`,
			user.Login,
			user.PasswordHash,
		).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserAlreadyExists
		}
		if err != nil {
			return fmt.Errorf("failed to create a new user %s: %w", user.Login, err)
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return userID, nil
}
