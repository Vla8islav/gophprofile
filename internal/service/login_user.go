package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/helpers"
)

func (m gophprofileService) LoginUser(ctx context.Context, userRegReq domain.UserLoginRequest) (*domain.AuthResult, error) {
	user, err := m.repository.GetUserByLogin(ctx, userRegReq.Login)
	if err != nil {
		return nil, fmt.Errorf("couldn't find user %s: %w", userRegReq.Login, err)
	}

	if !helpers.CheckPassword(userRegReq.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid user credentials for user %s: %w", userRegReq.Login, domain.ErrInvalidUserCredentials)
	}

	token, err := helpers.CreateAuthToken(user.ID, m.authSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token for user %d: %w", user.ID, err)
	}

	return &domain.AuthResult{
		Token:  token,
		UserID: user.ID,
	}, err
}
