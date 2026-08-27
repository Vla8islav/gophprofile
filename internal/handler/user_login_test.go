package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newTestLoginHandler(service domain.GophprofileService) *Handler {
	return &Handler{
		service: service,
		logger:  zap.NewNop(),
	}
}

func TestUserLoginHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{
			UserID: 123,
			Token:  "test-token",
		}, nil)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))

	// Token now comes back in the JSON body (not a cookie).
	var resp domain.UserLoginResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, "test-token", resp.Token)

	// And there should be NO auth cookie anymore.
	require.Empty(t, res.Cookies())
}

func TestUserLoginHandler_AllowsJSONContentTypeWithCharset(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{UserID: 123, Token: "test-token"}, nil)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUserLoginHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestLoginHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/user/login", nil)
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestUserLoginHandler_BadContentType(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_EmptyLogin(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_EmptyPassword(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_InvalidCredentials(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "wrong-password",
		}).
		Return(nil, repository.ErrUserNotFound)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"wrong-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestUserLoginHandler_ServiceError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(nil, errors.New("service error"))

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
