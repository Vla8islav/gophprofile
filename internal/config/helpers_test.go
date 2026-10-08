package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogSetEnv_RedactsSecrets(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	lg := zap.New(core)
	t.Setenv("DATABASE_URI", "postgres://user:supersecret@db:5432/x")
	t.Setenv("AUTH_TOKEN_SECRET", "tok-secret-value")
	t.Setenv("RUN_ADDRESS", ":8080")
	_, err := ReadFlagsServer([]string{}, lg)
	require.NoError(t, err)

	all := fmt.Sprint(logs.All())
	require.NotContains(t, all, "supersecret")
	require.NotContains(t, all, "tok-secret-value")
	require.Contains(t, all, ":8080") // non-secrets still logged
	require.Contains(t, all, "xxxxx") // DSN password masked, not dropped
}
