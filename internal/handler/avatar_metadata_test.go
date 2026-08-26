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

func TestAvatarMetadataHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetAvatarMetadata(gomock.Any(), testAvatarID).
		Return(&domain.Avatar{
			ID:               testAvatarID,
			UserID:           42,
			FileName:         "cat.png",
			MimeType:         "image/png",
			SizeBytes:        1024,
			ProcessingStatus: domain.ProcessingStatusCompleted,
			ThumbnailS3Keys:  map[string]string{"100x100": "thumbnails/x/100x100.jpg"},
		}, nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID+"/metadata", nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarMetadataHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	var resp domain.AvatarMetadataResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, testAvatarID, resp.ID)
	require.Equal(t, "cat.png", resp.FileName)
	require.Equal(t, int64(1024), resp.Size)
	require.Len(t, resp.Thumbnails, 1)
	require.Equal(t, "100x100", resp.Thumbnails[0].Size)
	require.Equal(t, "/api/v1/avatars/"+testAvatarID+"?size=100x100", resp.Thumbnails[0].URL)
}

func TestAvatarMetadataHandler_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		GetAvatarMetadata(gomock.Any(), testAvatarID).
		Return(nil, domain.ErrAvatarNotFound)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+testAvatarID+"/metadata", nil)
	req = withChiParams(req, map[string]string{"avatar_id": testAvatarID})
	w := httptest.NewRecorder()

	h.AvatarMetadataHandler(w, req)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestUserAvatarsListHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		ListUserAvatars(gomock.Any(), int64(42)).
		Return([]domain.Avatar{
			{ID: "a1", UserID: 42, FileName: "one.png"},
			{ID: "a2", UserID: 42, FileName: "two.jpg"},
		}, nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42/avatars", nil)
	req = withChiParams(req, map[string]string{"user_id": "42"})
	w := httptest.NewRecorder()

	h.UserAvatarsListHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	var resp domain.AvatarListResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Len(t, resp.Avatars, 2)
	require.Equal(t, "one.png", resp.Avatars[0].FileName)
}

func TestUserAvatarsListHandler_EmptyListIsEmptyArray(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		ListUserAvatars(gomock.Any(), int64(42)).
		Return(nil, nil)

	h := newAvatarTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42/avatars", nil)
	req = withChiParams(req, map[string]string{"user_id": "42"})
	w := httptest.NewRecorder()

	h.UserAvatarsListHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	// [] rather than null
	require.JSONEq(t, `{"avatars":[]}`, w.Body.String())
}
