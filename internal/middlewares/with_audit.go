package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/Vla8islav/gophprofile/internal/audit"
)

// WithAudit records one audit event per request. must be the outermost middleware
func WithAudit(publisher *audit.Publisher) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := audit.NewRequestData()
			ctx := audit.WithRequestData(r.Context(), data)

			lrw := &loggingResponseWriter{ResponseWriter: w, responseData: &responseData{}}
			start := time.Now()

			next.ServeHTTP(lrw, r.WithContext(ctx))

			status := lrw.responseData.status
			if status == 0 {
				status = http.StatusOK // a Write with no explicit WriteHeader is 200
			}

			event := audit.Event{
				Time:       audit.UnixTime{Time: start},
				Operation:  data.Operation,
				UserID:     data.UserID,
				SecretID:   data.SecretID,
				RemoteAddr: clientIP(r),
				Status:     status,
			}
			// Best-effort: an audit sink failure must not break the response.
			_ = publisher.Publish(r.Context(), event)
		})
	}
}

// clientIP returns the real client IP honoring a reverse proxy's forwarding
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i]) // first entry = original client
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	return r.RemoteAddr
}
