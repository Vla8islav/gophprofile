package middlewares

import (
	"net/http"
	"time"

	"github.com/Vla8islav/gophprofile/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func WithRequestLogger(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logging.WithTrace(r.Context(), base) // trace_id/span_id
			ctx := logging.Into(r.Context(), l)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r.WithContext(ctx))

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			l.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("route", chi.RouteContext(r.Context()).RoutePattern()),
				zap.Int("status", status),
				zap.Int("size", ww.BytesWritten()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
