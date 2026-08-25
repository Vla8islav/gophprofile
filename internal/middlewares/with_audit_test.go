package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophprofile/internal/audit"
)

type captureSink struct{ events []audit.Event }

func (c *captureSink) Write(_ context.Context, e audit.Event) error {
	c.events = append(c.events, e)
	return nil
}

func TestWithAudit_RecordsAnnotatedEvent(t *testing.T) {
	sink := &captureSink{}
	pub := audit.NewPublisher(sink)

	// A handler that annotates the audit RequestData and returns 201.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		audit.SetOperation(r.Context(), "secret.create")
		audit.SetSecretID(r.Context(), "abc-123")
		audit.SetUserID(r.Context(), 42)
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/secret/create", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.1")
	WithAudit(pub)(h).ServeHTTP(httptest.NewRecorder(), req)

	require.Len(t, sink.events, 1)
	e := sink.events[0]
	require.Equal(t, "secret.create", e.Operation)
	require.Equal(t, "abc-123", e.SecretID)
	require.Equal(t, int64(42), e.UserID)
	require.Equal(t, http.StatusCreated, e.Status)
	require.Equal(t, "203.0.113.7", e.RemoteAddr) //< first X-Forwarded-For entry
}

func TestWithAudit_DefaultsStatusTo200(t *testing.T) {
	sink := &captureSink{}
	pub := audit.NewPublisher(sink)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		audit.SetOperation(r.Context(), "secret.list")
		_, _ = w.Write([]byte("[]"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/secret/list", nil)
	WithAudit(pub)(h).ServeHTTP(httptest.NewRecorder(), req)

	require.Len(t, sink.events, 1)
	require.Equal(t, http.StatusOK, sink.events[0].Status)
	require.Equal(t, "secret.list", sink.events[0].Operation)
}
