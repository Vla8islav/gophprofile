package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, cfg *config.OptionsServer) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Get("/health", h.HealthHandler)
	r.Get("/api/ping", h.DBPing)
	r.Post("/api/user/register", h.UserRegisterHandler)
	r.Post("/api/user/login", h.UserLoginHandler)

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
