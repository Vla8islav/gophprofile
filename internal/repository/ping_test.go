package repository

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPostgresStorage_Ping(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	err := storage.Ping(ctx)
	require.NoError(t, err)
}
