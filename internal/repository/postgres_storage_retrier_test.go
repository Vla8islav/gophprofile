package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// If you already have an equivalent storage-setup helper, reuse it and drop this.
func newSecretStorage(t *testing.T) (*PostgresStorage, context.Context) {
	t.Helper()
	cfg, err := config.ReadFlagsServer(nil, zap.NewNop())
	require.NoError(t, err)
	return InitTestPostgresStorage(t, cfg), context.Background()
}

// The commit path: multiple statements in one transaction all persist.
func TestWithRetryTx_CommitsAllStatements(t *testing.T) {
	storage, ctx := newSecretStorage(t)

	_, err := storage.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS tx_tmp_commit (n INT)`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = storage.db.ExecContext(context.Background(), `DROP TABLE IF EXISTS tx_tmp_commit`)
	})

	err = storage.withRetryTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tx_tmp_commit (n) VALUES (1)`); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tx_tmp_commit (n) VALUES (2)`); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	var count int
	require.NoError(t, storage.db.QueryRowContext(ctx,
		`SELECT count(*) FROM tx_tmp_commit`).Scan(&count))
	require.Equal(t, 2, count)
}

// The rollback path: if the closure errors, NOTHING it wrote persists.
func TestWithRetryTx_RollsBackOnError(t *testing.T) {
	storage, ctx := newSecretStorage(t)

	_, err := storage.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS tx_tmp_rollback (n INT)`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = storage.db.ExecContext(context.Background(), `DROP TABLE IF EXISTS tx_tmp_rollback`)
	})

	sentinel := errors.New("forced failure inside tx")
	err = storage.withRetryTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO tx_tmp_rollback (n) VALUES (1)`); err != nil {
			return err
		}
		return sentinel // non-retriable → one rollback, then returned
	})
	require.ErrorIs(t, err, sentinel)

	// The insert was rolled back - zero rows persist.
	var count int
	require.NoError(t, storage.db.QueryRowContext(ctx, `SELECT count(*) FROM tx_tmp_rollback`).Scan(&count))
	require.Zero(t, count)
}
