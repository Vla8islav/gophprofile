package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func WithLogging(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := zap.String("method", r.Method)
			uri := zap.String("path", r.URL.Path)
			lrw := &loggingResponseWriter{ResponseWriter: w,
				responseData: &responseData{}}

			start := time.Now()

			next.ServeHTTP(lrw, r)

			end := time.Now()
			elapsed := zap.Duration("elapsed time", end.Sub(start))
			logger.Info("Request ", method, uri, elapsed)
			logger.Info("Response ", zap.Int("size", lrw.responseData.size),
				zap.Int("status", lrw.responseData.status))
		})
	}
}
