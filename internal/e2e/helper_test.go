package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/filestorage"
	"github.com/Vla8islav/gophprofile/internal/handler"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/Vla8islav/gophprofile/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func startStack(t *testing.T) string {
	t.Helper()
	setupOnce.Do(func() {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		ctx := context.Background()

		pgContainer, err := tcpostgres.Run(ctx,
			"postgres:16-alpine",
			tcpostgres.WithDatabase("gophprofile_e2e"),
			tcpostgres.WithUsername("e2e"),
			tcpostgres.WithPassword("e2e"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			setupErr = fmt.Errorf("start postgres: %w", err)
			return
		}
		dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			setupErr = fmt.Errorf("postgres dsn: %w", err)
			return
		}

		minioContainer, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
		if err != nil {
			setupErr = fmt.Errorf("start minio: %w", err)
			return
		}
		s3Endpoint, err := minioContainer.ConnectionString(ctx)
		if err != nil {
			setupErr = fmt.Errorf("minio endpoint: %w", err)
			return
		}

		cfg, err := config.ReadFlagsServer(nil, zap.NewNop())
		if err != nil {
			setupErr = fmt.Errorf("config: %w", err)
			return
		}
		cfg.DatabaseURI.Value = dsn
		cfg.DatabaseURI.BeenSet = true
		cfg.AuthTokenSecret.Value = "e2e-test-secret"

		db, err := repository.NewPostgresStorage(cfg, "../../migrations")
		if err != nil {
			setupErr = fmt.Errorf("postgres storage: %w", err)
			return
		}
		fs, err := filestorage.NewMinioStorage(ctx, s3Endpoint,
			minioContainer.Username, minioContainer.Password, "avatars-e2e", false)
		if err != nil {
			setupErr = fmt.Errorf("minio storage: %w", err)
			return
		}

		svc := service.NewGophprofileService(db, fs, cfg.AuthTokenSecret.Value)
		h := handler.NewHandler(svc, zap.NewNop())
		server := httptest.NewServer(handler.NewRouter(h, cfg))
		baseURL = server.URL
		// No explicit teardown: the process exit reaps the httptest server,
		// and with Ryuk disabled the containers are removed by --rm/GC.
	})
	require.NoError(t, setupErr)
	return baseURL
}

func registerUser(t *testing.T, base, login string) (token string) {
	t.Helper()
	body := fmt.Sprintf(`{"login":%q,"password":"e2e-password-1"}`, login)
	res, err := http.Post(base+"/api/user/register", "application/json",
		bytes.NewBufferString(body))
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp domain.UserRegisterResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.NotEmpty(t, resp.Token)
	return resp.Token
}

func pngFixture(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func uploadAvatar(t *testing.T, base, token, fileName string, content []byte) (*http.Response, []byte) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, base+"/api/v1/avatars", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, respBody
}

func doRequest(t *testing.T, method, url, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return res
}
