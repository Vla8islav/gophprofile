package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Переиспользование одноразового запуска тестовой БД
var (
	testStorageOnce sync.Once
	testStorage     *PostgresStorage
	testStorageErr  error
	testPGContainer testcontainers.Container
)

func InitTestPostgresStorage(t *testing.T, cfg *config.OptionsServer) *PostgresStorage {
	t.Helper()
	testStorageOnce.Do(func() {
		SetupTestPostgres(t, cfg)
		testStorage, testStorageErr = NewPostgresStorage(cfg, "../../migrations")
	})
	require.NoError(t, testStorageErr)

	return testStorage
}

func SetupTestPostgres(t *testing.T, cfg *config.OptionsServer) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("gophprofile_test"),
		tcpostgres.WithUsername("default_user"),
		tcpostgres.WithPassword("default_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	testPGContainer = pgContainer

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	cfg.DatabaseURI.Value = dsn
	cfg.DatabaseURI.BeenSet = true
}
