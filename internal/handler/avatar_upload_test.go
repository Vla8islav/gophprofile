package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func pngFixture(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func multipartBody(t *testing.T, fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return body, writer.FormDataContentType()
}

func TestAvatarUploadHandler_Success(t *testing.T) {
	t.Parallel()
	pngBytes := pngFixture(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	created := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		UploadAvatar(gomock.Any(), int64(42), "cat.png", "image/png", int64(len(pngBytes)), gomock.Any()).
		DoAndReturn(func(_ any, _ int64, _, _ string, size int64, content io.Reader) (*domain.Avatar, error) {
			// The handler must hand over the full file even after sniffing magic bytes.
			data, err := io.ReadAll(content)
			require.NoError(t, err)
			require.Equal(t, pngBytes, data)
			return &domain.Avatar{
				ID:               "7c9e6679-7425-40de-944b-e07fc1f90ae7",
				UserID:           42,
				ProcessingStatus: domain.ProcessingStatusPending,
				CreatedAt:        created,
			}, nil
		})

	h := newAvatarTestHandler(service)

	body, contentType := multipartBody(t, "file", "cat.png", pngBytes)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", contentType)
	req = asUser(req, 42)
	w := httptest.NewRecorder()

	h.AvatarUploadHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	var resp domain.AvatarUploadResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, "7c9e6679-7425-40de-944b-e07fc1f90ae7", resp.ID)
	require.Equal(t, int64(42), resp.UserID)
	require.Equal(t, "/api/v1/avatars/7c9e6679-7425-40de-944b-e07fc1f90ae7", resp.URL)
	require.Equal(t, domain.ProcessingStatusPending, resp.Status)
}

func TestAvatarUploadHandler_Unauthenticated(t *testing.T) {
	t.Parallel()
	pngBytes := pngFixture(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	body, contentType := multipartBody(t, "file", "cat.png", pngBytes)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", contentType)
	// no asUser: context carries no user id
	w := httptest.NewRecorder()

	h.AvatarUploadHandler(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestAvatarUploadHandler_RejectsNonImage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	// text bytes named .png: magic-byte sniffing must reject it
	body, contentType := multipartBody(t, "file", "fake.png", []byte("just some text pretending"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", contentType)
	req = asUser(req, 42)
	w := httptest.NewRecorder()

	h.AvatarUploadHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
	var apiErr apiError
	require.NoError(t, json.NewDecoder(res.Body).Decode(&apiErr))
	require.Equal(t, "Invalid file format", apiErr.Error)
	require.Contains(t, apiErr.Details, "jpeg, png, webp")
}

func TestAvatarUploadHandler_MissingFileField(t *testing.T) {
	t.Parallel()
	pngBytes := pngFixture(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	body, contentType := multipartBody(t, "not_file", "cat.png", pngBytes)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", contentType)
	req = asUser(req, 42)
	w := httptest.NewRecorder()

	h.AvatarUploadHandler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestAvatarUploadHandler_FileTooLarge(t *testing.T) {
	t.Parallel()
	pngBytes := pngFixture(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	big := make([]byte, domain.MaxAvatarSizeBytes+1)
	copy(big, pngBytes)
	body, contentType := multipartBody(t, "file", "huge.png", big)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", contentType)
	req = asUser(req, 42)
	w := httptest.NewRecorder()

	h.AvatarUploadHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)
	var apiErr apiError
	require.NoError(t, json.NewDecoder(res.Body).Decode(&apiErr))
	require.Equal(t, "File too large", apiErr.Error)
	require.Equal(t, int64(domain.MaxAvatarSizeBytes), apiErr.MaxSize)
}
