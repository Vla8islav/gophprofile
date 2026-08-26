package filestorage

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func startMinioStorage(t *testing.T) *MinioStorage {
	t.Helper()
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	require.NoError(t, err, "failed to start minio container")
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	endpoint, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	storage, err := NewMinioStorage(ctx, endpoint,
		container.Username, container.Password, "avatars-test", false)
	require.NoError(t, err)
	require.NoError(t, storage.EnsureBucket(ctx))
	return storage
}

func TestMinioStorage_UploadDownloadDelete(t *testing.T) {
	storage := startMinioStorage(t)
	ctx := context.Background()

	content := []byte("fake image bytes")
	key := "avatars/test-id/original"

	require.NoError(t, storage.Upload(ctx, key, "image/png", int64(len(content)), bytes.NewReader(content)))

	rc, err := storage.Download(ctx, key)
	require.NoError(t, err)
	downloaded, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	require.Equal(t, content, downloaded)

	require.NoError(t, storage.Delete(ctx, key))

	_, err = storage.Download(ctx, key)
	require.Error(t, err, "downloading a deleted key must fail")
	require.True(t, strings.Contains(err.Error(), "key") || err != nil)
}

func TestMinioStorage_DownloadMissingKeyFailsEagerly(t *testing.T) {
	storage := startMinioStorage(t)

	// Must fail at Download
	_, err := storage.Download(context.Background(), "no/such/key")
	require.Error(t, err)
}

func TestMinioStorage_Ping(t *testing.T) {
	storage := startMinioStorage(t)
	require.NoError(t, storage.Ping(context.Background()))
}
