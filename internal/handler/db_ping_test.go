package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newTestHandler(t *testing.T) (*Handler, *mocks.MockGophkeeperService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	service := mocks.NewMockGophkeeperService(ctrl)
	h := &Handler{service: service, logger: zap.NewNop()}
	return h, service
}

func validCreateBody(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(domain.CreateUserParams{
		ID:      uuid.New(),
		Type:    domain.SecretTypeText,
		Payload: []byte("cipher"),
	})
	require.NoError(t, err)
	return b
}

func TestDBPing_OK(t *testing.T) {
	h, service := newTestHandler(t)
	service.EXPECT().Ping(gomock.Any()).Return(nil)

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodGet, "/api/ping", nil))
	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestDBPing_Error(t *testing.T) {
	h, service := newTestHandler(t)
	service.EXPECT().Ping(gomock.Any()).Return(errors.New("db down"))

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodGet, "/api/ping", nil))
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestDBPing_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodPost, "/api/ping", nil))
	require.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}
