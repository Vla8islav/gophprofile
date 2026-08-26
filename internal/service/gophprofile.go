package service

import (
	"context"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.uber.org/zap"
)

type gophprofileService struct {
	repository  domain.GophprofileRepository
	fileStorage domain.FileStorage
	events      domain.EventPublisher
	logger      *zap.Logger
	authSecret  []byte
}

func NewGophprofileService(
	repo domain.GophprofileRepository,
	fileStorage domain.FileStorage,
	events domain.EventPublisher,
	logger *zap.Logger,
	authSecret string,
) domain.GophprofileService {
	return gophprofileService{
		repository:  repo,
		fileStorage: fileStorage,
		events:      events,
		logger:      logger,
		authSecret:  []byte(authSecret),
	}
}

// publishEvent sends an event without failing the caller's operation
func (m gophprofileService) publishEvent(ctx context.Context, key, eventType string, payload any) {
	if err := m.events.Publish(ctx, key, eventType, payload); err != nil {
		m.logger.Error("failed to publish event",
			zap.String("event_type", eventType),
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

func (m gophprofileService) BrokerPing(ctx context.Context) error {
	return m.events.Ping(ctx)
}
