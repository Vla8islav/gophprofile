package service

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
)

func (m gophprofileService) CreateUser(ctx context.Context, userRegReq domain.UserRegisterRequest) (*domain.AuthResult, error) {
	hash, err := helpers.HashPassword(userRegReq.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate the hash for the new user %s: %w",
			userRegReq.Login, err)
	}

	salt := make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate kdf salt for user %s: %w",
			userRegReq.Login, err)
	}

	createUserParams := domain.CreateUserParams{
		Login:        userRegReq.Login,
		PasswordHash: hash,
		Salt:         salt,
	}
	userID, err := m.repository.CreateUser(ctx, createUserParams)

	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	token, err := helpers.CreateAuthToken(userID, m.authSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token for user %d: %w", userID, err)
	}

	return &domain.AuthResult{
		Token:  token,
		UserID: userID,
	}, err
}
