package handler

import (
	"bytes"
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

func newTestRegisterHandler(service domain.GophprofileService) *Handler {
	return &Handler{
		service: service,
		logger:  zap.NewNop(),
	}
}

func TestUserRegisterHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		CreateUser(gomock.Any(), domain.UserRegisterRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{
			Token:  "test-token",
			UserID: 1,
		}, nil)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUserRegisterHandler_AllowsJSONContentTypeWithCharset(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		CreateUser(gomock.Any(), domain.UserRegisterRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{
			Token:  "test-token",
			UserID: 1,
		}, nil)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUserRegisterHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestRegisterHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/user/register", nil)
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestUserRegisterHandler_BadContentType(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserRegisterHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserRegisterHandler_EmptyPassword(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserRegisterHandler_ServiceError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		CreateUser(gomock.Any(), domain.UserRegisterRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(nil, errors.New("service error"))

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestUserRegisterHandler_UserAlreadyExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophprofileService(ctrl)
	service.EXPECT().
		CreateUser(gomock.Any(), domain.UserRegisterRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(nil, repository.ErrUserAlreadyExists)

	h := newTestRegisterHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserRegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusConflict, res.StatusCode)
}
