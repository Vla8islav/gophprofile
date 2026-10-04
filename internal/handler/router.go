package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, cfg *config.OptionsServer) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.ClientIPFromXFFTrustedProxies(1))
	r.Use(middleware.StripSlashes)
	r.Use(otelchi.Middleware("gophprofile-server", otelchi.WithChiRoutes(r),
		otelchi.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/metrics" &&
				r.URL.Path != "/health" &&
				!strings.HasPrefix(r.URL.Path, "/web/static/") // don't trace garbage requests
		}))) // creates the span
	r.Use(middlewares.WithRequestLogger(h.logger)) // reads the span
	r.Use(middlewares.WithMetrics)

	rps, err := strconv.ParseFloat(cfg.RateLimitRPS.Value, 64)
	if err != nil || rps <= 0 {
		rps = 10
	}
	burst, err := strconv.Atoi(cfg.RateLimitBurst.Value)
	if err != nil || burst <= 0 {
		burst = 20
	}
	r.Use(middlewares.WithRateLimit(rps, burst))

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/health", h.HealthHandler)
	r.Get("/api/ping", h.DBPing)
	r.Post("/api/user/register", h.UserRegisterHandler)
	r.Post("/api/user/login", h.UserLoginHandler)

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	// web interface
	r.Get("/web", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/upload", http.StatusMovedPermanently)
	})
	r.Get("/web/upload", h.serveWebFile("index.html",
		"text/html; charset=utf-8"))
	r.Get("/web/static/auth.js", h.serveWebFile("auth.js",
		"application/javascript; charset=utf-8"))

	// Public read endpoints
	r.Get("/api/v1/avatars/{avatar_id}", h.AvatarGetHandler)
	r.Get("/api/v1/avatars/{avatar_id}/metadata", h.AvatarMetadataHandler)
	r.Get("/api/v1/users/{user_id}/avatar", h.UserAvatarGetHandler)
	r.Get("/api/v1/users/{user_id}/avatars", h.UserAvatarsListHandler)

	// Mutating endpoints require a Bearer token
	r.Group(func(r chi.Router) {
		r.Use(middlewares.WithAuth([]byte(cfg.AuthTokenSecret.Value)))

		r.Post("/api/v1/avatars", h.AvatarUploadHandler)
		r.Delete("/api/v1/avatars/{avatar_id}", h.AvatarDeleteHandler)
		r.Delete("/api/v1/users/{user_id}/avatar", h.UserAvatarDeleteHandler)
	})

	return r
}
