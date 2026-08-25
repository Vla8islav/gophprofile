package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWithLogging_PassesThrough(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hi"))
	})

	rec := httptest.NewRecorder()
	WithLogging(zap.NewNop())(h).ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/x", nil))

	require.True(t, called)
	require.Equal(t, http.StatusTeapot, rec.Result().StatusCode)
}

func TestChainMiddlewares_OutermostFirst(t *testing.T) {
	var order []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	// first arg is the outermost wrapper
	ChainMiddlewares(final, mw("a"), mw("b")).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, []string{"a", "b", "handler"}, order)
}
