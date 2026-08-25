package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresStorage struct {
	config     *config.OptionsServer
	db         *sql.DB
	classifier *PostgresErrorClassifier
}

func (s *PostgresStorage) isRetriablePostgresError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if s.classifier == nil {
		return false
	}

	switch s.classifier.Classify(pgErr) {
	case Retriable:
		return true
	case NonRetriable:
		return false
	default:
		return false
	}
}

func NewPostgresStorage(config *config.OptionsServer, migrationsFolder string) (*PostgresStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if !config.DatabaseURI.BeenSet {
		return nil, fmt.Errorf("database DSN wasn't set")
	}

	dsn := config.DatabaseURI.Value
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	storage := PostgresStorage{config: config, db: db, classifier: NewPostgresErrorClassifier()}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = storage.Ping(ctx)
	if err != nil {
		if closeError := storage.db.Close(); closeError != nil {
			return nil, fmt.Errorf("failed to ping postgres %w also failed to close db %v", err, closeError)
		}
		return nil, fmt.Errorf("failed to ping postgres %w", err)
	}

	// Run all pending migrations from migrations/
	if err := goose.Up(db, migrationsFolder); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply goose migrations: %w", err)
	}

	return &storage, nil
}
