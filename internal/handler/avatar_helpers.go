package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// apiError is the JSON error envelope the avatar API uses (per spec):
// {"error": "...", "details": "...", "max_size": ...}
type apiError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	MaxSize int64  `json:"max_size,omitempty"`
}

func (h *Handler) writeJSONError(w http.ResponseWriter, status int, apiErr apiError) {
	h.logger.Info("api error",
		zap.Int("status", status),
		zap.String("error", apiErr.Error),
		zap.String("details", apiErr.Details),
	)
	h.writeJSON(w, status, apiErr)
}

func (h *Handler) writeAvatarNotFound(w http.ResponseWriter) {
	h.writeJSONError(w, http.StatusNotFound, apiError{Error: "Avatar not found"})
}

// requestUserID extracts the authenticated user from the request context
func (h *Handler) requestUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorised(w, "authentication required")
		return 0, false
	}
	return userID, true
}

// avatarIDParam validates the {avatar_id} path param. A malformed UUID cannot
// match any avatar, so it answers 404 (not 400) just like a well-formed
// unknown id would.
func (h *Handler) avatarIDParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	avatarID := chi.URLParam(r, "avatar_id")
	if _, err := uuid.Parse(avatarID); err != nil {
		h.writeAvatarNotFound(w)
		return "", false
	}
	return avatarID, true
}

func (h *Handler) userIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, apiError{
			Error:   "Invalid user id",
			Details: "user_id must be an integer",
		})
		return 0, false
	}
	return userID, true
}

var allowedSizeVariants = map[string]bool{
	"":         true,
	"original": true,
	"100x100":  true,
	"300x300":  true,
}

func (h *Handler) sizeParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	size := r.URL.Query().Get("size")
	if !allowedSizeVariants[size] {
		h.writeJSONError(w, http.StatusBadRequest, apiError{
			Error:   "Invalid size",
			Details: "Supported sizes: 100x100, 300x300, original",
		})
		return "", false
	}
	return size, true
}

func avatarETag(avatar *domain.Avatar) string {
	return fmt.Sprintf(`"%s-%d"`, avatar.ID, avatar.UpdatedAt.Unix())
}

func (h *Handler) serveAvatarContent(w http.ResponseWriter, r *http.Request, avatar *domain.Avatar, content io.ReadCloser) {
	defer func() { _ = content.Close() }()

	etag := avatarETag(avatar)
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", avatar.MimeType)
	if _, err := io.Copy(w, content); err != nil {
		// Headers are already out; all we can do is log the broken transfer.
		h.logger.Warn("failed to stream avatar content", zap.Error(err))
	}
}
