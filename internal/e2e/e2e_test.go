// Package e2e taht wires the real router, service, repository, Postgres and MinIO
package e2e

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/stretchr/testify/require"
)

var (
	setupOnce sync.Once
	baseURL   string
	setupErr  error
)

func TestE2E_Health(t *testing.T) {
	base := startStack(t)

	res, err := http.Get(base + "/health")
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	var health struct {
		Status     string            `json:"status"`
		Components map[string]string `json:"components"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&health))
	require.Equal(t, "ok", health.Status)
	require.Equal(t, "ok", health.Components["database"])
	require.Equal(t, "ok", health.Components["s3"])
}

func TestE2E_AvatarLifecycle(t *testing.T) {
	base := startStack(t)
	token := registerUser(t, base, "lifecycle-user")
	original := pngFixture(t)

	// Upload
	res, body := uploadAvatar(t, base, token, "cat.png", original)
	require.Equal(t, http.StatusCreated, res.StatusCode, string(body))
	var uploaded domain.AvatarUploadResponse
	require.NoError(t, json.Unmarshal(body, &uploaded))
	require.NotEmpty(t, uploaded.ID)
	require.Equal(t, base+"/api/v1/avatars/"+uploaded.ID, base+uploaded.URL)

	// Download: identical bytes
	getRes := doRequest(t, http.MethodGet, base+uploaded.URL, "")
	defer getRes.Body.Close()
	require.Equal(t, http.StatusOK, getRes.StatusCode)
	require.Equal(t, "image/png", getRes.Header.Get("Content-Type"))
	require.NotEmpty(t, getRes.Header.Get("ETag"))
	downloaded, err := io.ReadAll(getRes.Body)
	require.NoError(t, err)
	require.Equal(t, original, downloaded)

	// Conditional GET honours the ETag
	req, err := http.NewRequest(http.MethodGet, base+uploaded.URL, nil)
	require.NoError(t, err)
	req.Header.Set("If-None-Match", getRes.Header.Get("ETag"))
	condRes, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer condRes.Body.Close()
	require.Equal(t, http.StatusNotModified, condRes.StatusCode)

	// Metadata
	metaRes := doRequest(t, http.MethodGet, base+uploaded.URL+"/metadata", "")
	defer metaRes.Body.Close()
	require.Equal(t, http.StatusOK, metaRes.StatusCode)
	var meta domain.AvatarMetadataResponse
	require.NoError(t, json.NewDecoder(metaRes.Body).Decode(&meta))
	require.Equal(t, "cat.png", meta.FileName)
	require.Equal(t, int64(len(original)), meta.Size)

	// User-scoped routes see it
	userAvatarRes := doRequest(t, http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%d/avatar", base, uploaded.UserID), "")
	defer userAvatarRes.Body.Close()
	require.Equal(t, http.StatusOK, userAvatarRes.StatusCode)

	listRes := doRequest(t, http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%d/avatars", base, uploaded.UserID), "")
	defer listRes.Body.Close()
	var list domain.AvatarListResponse
	require.NoError(t, json.NewDecoder(listRes.Body).Decode(&list))
	require.Len(t, list.Avatars, 1)

	// Delete, then everything 404s
	delRes := doRequest(t, http.MethodDelete, base+uploaded.URL, token)
	delRes.Body.Close()
	require.Equal(t, http.StatusNoContent, delRes.StatusCode)

	goneRes := doRequest(t, http.MethodGet, base+uploaded.URL, "")
	goneRes.Body.Close()
	require.Equal(t, http.StatusNotFound, goneRes.StatusCode)
}

func TestE2E_AuthEnforcement(t *testing.T) {
	base := startStack(t)
	ownerToken := registerUser(t, base, "auth-owner")
	strangerToken := registerUser(t, base, "auth-stranger")

	res, body := uploadAvatar(t, base, ownerToken, "cat.png", pngFixture(t))
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var uploaded domain.AvatarUploadResponse
	require.NoError(t, json.Unmarshal(body, &uploaded))

	// No token: 401
	noAuthUpload, _ := uploadAvatar(t, base, "", "cat.png", pngFixture(t))
	require.Equal(t, http.StatusUnauthorized, noAuthUpload.StatusCode)

	noAuthDel := doRequest(t, http.MethodDelete, base+uploaded.URL, "")
	noAuthDel.Body.Close()
	require.Equal(t, http.StatusUnauthorized, noAuthDel.StatusCode)

	// Someone else's token: 403, and the avatar survives
	strangerDel := doRequest(t, http.MethodDelete, base+uploaded.URL, strangerToken)
	strangerDel.Body.Close()
	require.Equal(t, http.StatusForbidden, strangerDel.StatusCode)

	stillThere := doRequest(t, http.MethodGet, base+uploaded.URL, "")
	stillThere.Body.Close()
	require.Equal(t, http.StatusOK, stillThere.StatusCode)
}

func TestE2E_UploadValidation(t *testing.T) {
	base := startStack(t)
	token := registerUser(t, base, "validation-user")

	// Text masquerading as PNG
	res, body := uploadAvatar(t, base, token, "fake.png", []byte("plain text, no image here"))
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	require.Contains(t, string(body), "Invalid file format")

	// Over the 10MB cap: 413
	big := make([]byte, domain.MaxAvatarSizeBytes+1)
	copy(big, pngFixture(t))
	res, body = uploadAvatar(t, base, token, "huge.png", big)
	require.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)
	require.Contains(t, string(body), "File too large")
}

func TestE2E_NotFoundCases(t *testing.T) {
	base := startStack(t)

	res := doRequest(t, http.MethodGet,
		base+"/api/v1/avatars/7c9e6679-7425-40de-944b-e07fc1f90ae7", "")
	res.Body.Close()
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	res = doRequest(t, http.MethodGet, base+"/api/v1/avatars/not-a-uuid", "")
	res.Body.Close()
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	res = doRequest(t, http.MethodGet, base+"/api/v1/users/999999/avatar", "")
	res.Body.Close()
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	res = doRequest(t, http.MethodGet, base+"/api/v1/users/999999/avatars", "")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	var list domain.AvatarListResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
	require.Empty(t, list.Avatars)
}

// TestE2E_AsyncThumbnailPipeline drives the full outbox chain
func TestE2E_AsyncThumbnailPipeline(t *testing.T) {
	base := startStack(t)
	token := registerUser(t, base, "pipeline-user")

	res, body := uploadAvatar(t, base, token, "cat.png", pngFixture(t))
	require.Equal(t, http.StatusCreated, res.StatusCode, string(body))
	var uploaded domain.AvatarUploadResponse
	require.NoError(t, json.Unmarshal(body, &uploaded))
	require.Equal(t, "pending", uploaded.Status)

	var meta domain.AvatarMetadataResponse
	require.Eventually(t, func() bool {
		metaRes := doRequest(t, http.MethodGet, base+uploaded.URL+"/metadata", "")
		defer metaRes.Body.Close()
		if metaRes.StatusCode != http.StatusOK {
			return false
		}
		if err := json.NewDecoder(metaRes.Body).Decode(&meta); err != nil {
			return false
		}
		return meta.Status == "completed"
	}, 60*time.Second, 250*time.Millisecond, "avatar never reached completed status")

	require.Len(t, meta.Thumbnails, 2)

	thumbRes := doRequest(t, http.MethodGet, base+uploaded.URL+"?size=100x100", "")
	defer thumbRes.Body.Close()
	require.Equal(t, http.StatusOK, thumbRes.StatusCode)
	require.Equal(t, "image/jpeg", thumbRes.Header.Get("Content-Type"))
	thumb, _, err := image.Decode(thumbRes.Body)
	require.NoError(t, err)
	require.Equal(t, 100, thumb.Bounds().Dx())
	require.Equal(t, 100, thumb.Bounds().Dy())
}
