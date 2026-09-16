package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithRequestLogger_InjectsLoggerAndWritesAccessLine(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	base := zap.New(core)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.From(r.Context()).Info("from handler") // uses the injected logger
		w.WriteHeader(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	WithRequestLogger(base)(h).ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/x", nil))

	require.Equal(t, http.StatusNoContent, rec.Code)

	require.Equal(t, 1, logs.FilterMessage("from handler").Len())

	access := logs.FilterMessage("http request")
	require.Equal(t, 1, access.Len())
	fields := access.All()[0].ContextMap()
	require.Equal(t, "GET", fields["method"])
	require.EqualValues(t, http.StatusNoContent, fields["status"])
}
