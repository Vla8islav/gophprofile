package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAvatarDeleteHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		DeleteAvatar(gomock.Any(), testAvatarID, int64(42)).
		Return(nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+testAvatarID, nil)
	req = asUser(req, 42)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarDeleteHandler(w, req)

	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestAvatarDeleteHandler_ForbiddenForForeignAvatar(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		DeleteAvatar(gomock.Any(), testAvatarID, int64(43)).
		Return(domain.ErrNotAvatarOwner)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+testAvatarID, nil)
	req = asUser(req, 43)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarDeleteHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusForbidden, res.StatusCode)
	var apiErr apiError
	require.NoError(t, json.NewDecoder(res.Body).Decode(&apiErr))
	require.Equal(t, "Forbidden", apiErr.Error)
	require.Equal(t, "You can only delete your own avatars", apiErr.Details)
}

func TestAvatarDeleteHandler_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		DeleteAvatar(gomock.Any(), testAvatarID, int64(42)).
		Return(domain.ErrAvatarNotFound)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+testAvatarID, nil)
	req = asUser(req, 42)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarDeleteHandler(w, req)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestAvatarDeleteHandler_Unauthenticated(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := newAvatarTestHandler(mocks.NewMockGophprofileService(ctrl))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+testAvatarID, nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarDeleteHandler(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestUserAvatarDeleteHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		DeleteUserAvatar(gomock.Any(), int64(42), int64(42)).
		Return(nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/42/avatar", nil)
	req = asUser(req, 42)
	req = withChiParams(req, map[string]string{"user_id": "42"})
	w := httptest.NewRecorder()

	h.UserAvatarDeleteHandler(w, req)

	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}
