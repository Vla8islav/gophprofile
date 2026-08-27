package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const testAvatarID = "7c9e6679-7425-40de-944b-e07fc1f90ae7"

func TestAvatarGetHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	updated := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetAvatarContent(gomock.Any(), testAvatarID, "").
		Return(
			&domain.Avatar{ID: testAvatarID, MimeType: "image/png", UpdatedAt: updated},
			io.NopCloser(strings.NewReader("png-bytes")),
			nil,
		)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID, nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarGetHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "image/png", res.Header.Get("Content-Type"))
	require.Equal(t, "max-age=86400", res.Header.Get("Cache-Control"))
	require.NotEmpty(t, res.Header.Get("ETag"))

	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "png-bytes", string(data))
}

func TestAvatarGetHandler_ETagNotModified(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	updated := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	avatar := &domain.Avatar{ID: testAvatarID, MimeType: "image/png", UpdatedAt: updated}
	etag := avatarETag(avatar)

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetAvatarContent(gomock.Any(), testAvatarID, "").
		Return(avatar, io.NopCloser(strings.NewReader("png-bytes")), nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID, nil)
	req.Header.Set("If-None-Match", etag)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarGetHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotModified, res.StatusCode)
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Empty(t, data)
}

func TestAvatarGetHandler_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetAvatarContent(gomock.Any(), testAvatarID, "").
		Return(nil, nil, domain.ErrAvatarNotFound)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID, nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarGetHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode)
	var apiErr apiError
	require.NoError(t, json.NewDecoder(res.Body).Decode(&apiErr))
	require.Equal(t, "Avatar not found", apiErr.Error)
}

func TestAvatarGetHandler_MalformedUUIDIs404(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Service must not be called at all.
	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/not-a-uuid", nil)
	req = withChiParams(req, map[string]string{"avatar_id": "not-a-uuid"})
	w := httptest.NewRecorder()

	h.AvatarGetHandler(w, req)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestAvatarGetHandler_InvalidSizeParam(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID+"?size=999x999", nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarGetHandler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestUserAvatarGetHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetUserAvatarContent(gomock.Any(), int64(42), "").
		Return(
			&domain.Avatar{ID: testAvatarID, UserID: 42, MimeType: "image/jpeg"},
			io.NopCloser(strings.NewReader("jpeg-bytes")),
			nil,
		)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42/avatar", nil)
	req = withChiParams(req, map[string]string{"user_id": "42"})
	w := httptest.NewRecorder()

	h.UserAvatarGetHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "image/jpeg", res.Header.Get("Content-Type"))
}

func TestUserAvatarGetHandler_BadUserID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc/avatar", nil)
	req = withChiParams(req, map[string]string{"user_id": "abc"})
	w := httptest.NewRecorder()

	h.UserAvatarGetHandler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}
