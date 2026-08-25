package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User

	err := s.withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx,
			`SELECT id, login, password_hash FROM users WHERE login = $1`,
			login,
		).Scan(&user.ID, &user.Login, &user.PasswordHash)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get a user by login %s: %w", user.Login, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}
