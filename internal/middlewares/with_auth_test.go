package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithAuth(t *testing.T) {
	secret := []byte("test-secret")

	validToken, err := helpers.CreateAuthToken(42, secret)
	require.NoError(t, err)

	wrongSecretToken, err := helpers.CreateAuthToken(42, []byte("some-other-secret"))
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string // "" means header absent
		setHeader      bool
		wantStatus     int
		wantNextCalled bool
		wantUserID     int64
	}{
		{
			name:           "valid bearer token passes through",
			authHeader:     "Bearer " + validToken,
			setHeader:      true,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantUserID:     42,
		},
		{
			name:           "missing authorization header is rejected",
			setHeader:      false,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "missing Bearer prefix is rejected",
			authHeader:     validToken, // no "Bearer " prefix
			setHeader:      true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "Bearer prefix with empty token is rejected",
			authHeader:     "Bearer ",
			setHeader:      true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "garbage token is rejected",
			authHeader:     "Bearer not-a-real-token",
			setHeader:      true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "token signed with wrong secret is rejected",
			authHeader:     "Bearer " + wrongSecretToken,
			setHeader:      true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				nextCalled bool
				gotUserID  int64
			)

			// The protected handler. It should ONLY run when auth succeeds,
			// and when it runs, the user ID must be in the context.
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := WithAuth(secret)(next)

			req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
			if tt.setHeader {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "status code")
			assert.Equal(t, tt.wantNextCalled, nextCalled, "next handler invocation")
			if tt.wantNextCalled {
				assert.Equal(t, tt.wantUserID, gotUserID, "user id injected into context")
			}
		})
	}
}
