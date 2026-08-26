package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHealthHandler_AllOK(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().Ping(gomock.Any()).Return(nil)
	service.EXPECT().FileStoragePing(gomock.Any()).Return(nil)

	h := newAvatarTestHandler(service)

	w := httptest.NewRecorder()
	h.HealthHandler(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	var resp healthResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, "ok", resp.Status)
	require.Equal(t, "ok", resp.Components["database"])
	require.Equal(t, "ok", resp.Components["s3"])
}

func TestHealthHandler_S3Down(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().Ping(gomock.Any()).Return(nil)
	service.EXPECT().FileStoragePing(gomock.Any()).Return(errors.New("connection refused"))

	h := newAvatarTestHandler(service)

	w := httptest.NewRecorder()
	h.HealthHandler(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	var resp healthResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, "degraded", resp.Status)
	require.Equal(t, "ok", resp.Components["database"])
	require.Contains(t, resp.Components["s3"], "connection refused")
}
