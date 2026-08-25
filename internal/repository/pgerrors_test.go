package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestPostgresErrorClassifier(t *testing.T) {
	c := NewPostgresErrorClassifier()

	// nil and non-pg errors are non-retriable.
	require.Equal(t, NonRetriable, c.Classify(nil))
	require.Equal(t, NonRetriable, c.Classify(errors.New("plain error")))

	// Transient conditions are retriable.
	for _, code := range []string{
		pgerrcode.ConnectionFailure,
		pgerrcode.ConnectionException,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.CannotConnectNow,
	} {
		require.Equal(t, Retriable, c.Classify(&pgconn.PgError{Code: code}), "code %s should be retriable", code)
	}

	// Data/integrity/syntax errors are non-retriable.
	for _, code := range []string{
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedTable,
		"99999", // unknown code -> default non-retriable
	} {
		require.Equal(t, NonRetriable, c.Classify(&pgconn.PgError{Code: code}), "code %s should be non-retriable", code)
	}
}
