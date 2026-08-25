package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/google/uuid"
)

func (m gophprofileService) UpdateSecret(ctx context.Context, userID int64, id uuid.UUID, req domain.UpdateSecretRequest) (int64, error) {
	if id == uuid.Nil {
		return 0, fmt.Errorf("secret id is required: %w", domain.ErrInvalidSecretID)
	}

	params := domain.UpdateSecretParams{
		ID:      id,          // from the URL
		UserID:  userID,      // from the token
		Payload: req.Payload, // from the body
		Meta:    req.Meta,
		Version: req.Version,
	}

	newVersion, err := m.repository.UpdateSecret(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to update secret: %w", err)
	}

	return newVersion, nil
}
