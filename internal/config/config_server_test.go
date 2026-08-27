package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReadFlagsServer_Defaults(t *testing.T) {
	cfg, err := ReadFlagsServer(nil, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "localhost:8080", cfg.ServerAddress.Value)
	require.Equal(t, "./migrations", cfg.MigrationsFolder.Value)
	require.Equal(t, "", cfg.AuditLogPath.Value)
	require.False(t, cfg.ServerAddress.BeenSet)
}

func TestReadFlagsServer_Flags(t *testing.T) {
	cfg, err := ReadFlagsServer(
		[]string{"-a", "0.0.0.0:9999", "-d", "flag-dsn", "-s", "sekret", "-audit-log", "/tmp/a.jsonl"},
		zap.NewNop(),
	)
	require.NoError(t, err)
	require.Equal(t, "0.0.0.0:9999", cfg.ServerAddress.Value)
	require.True(t, cfg.ServerAddress.BeenSet)
	require.Equal(t, "flag-dsn", cfg.DatabaseURI.Value)
	require.Equal(t, "sekret", cfg.AuthTokenSecret.Value)
	require.Equal(t, "/tmp/a.jsonl", cfg.AuditLogPath.Value)
}

func TestReadFlagsServer_EnvOverridesFlags(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "env-host:1")
	t.Setenv("AUTH_TOKEN_SECRET", "env-secret")

	cfg, err := ReadFlagsServer([]string{"-a", "flag-host:2", "-s", "flag-secret"}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "env-host:1", cfg.ServerAddress.Value) // env beats flag
	require.Equal(t, "env-secret", cfg.AuthTokenSecret.Value)
}

func TestReadFlagsServer_ConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	require.NoError(t, os.WriteFile(path,
		[]byte(`{"server_address":"file-host:3","database_uri":"file-dsn"}`), 0o600))

	cfg, err := ReadFlagsServer([]string{"-config", path}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "file-host:3", cfg.ServerAddress.Value)
	require.Equal(t, "file-dsn", cfg.DatabaseURI.Value)
}

func TestReadFlagsServer_FlagBeatsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"server_address":"file-host:3"}`), 0o600))

	cfg, err := ReadFlagsServer([]string{"-config", path, "-a", "flag-host:4"}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "flag-host:4", cfg.ServerAddress.Value) // flag beats file
}

func TestReadFlagsServer_InvalidConfigFile(t *testing.T) {
	_, err := ReadFlagsServer([]string{"-config", "/no/such/file.json"}, zap.NewNop())
	require.Error(t, err)
}

func TestReadFlagsServer_UnknownFlagErrors(t *testing.T) {
	_, err := ReadFlagsServer([]string{"-totally-unknown-flag", "x"}, zap.NewNop())
	require.Error(t, err)
}
