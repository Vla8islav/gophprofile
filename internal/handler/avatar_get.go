package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// AvatarGetHandler godoc
// @Summary  Download an avatar (original or thumbnail)
// @Tags     avatars
// @Produce  image/jpeg,image/png,image/webp
// @Param    avatar_id path string true "avatar UUID"
// @Param    size query string false "100x100, 300x300 or original"
// @Success  200
// @Failure  400 {object} handler.apiError
// @Failure  404 {object} handler.apiError
// @Failure  500
// @Router   /api/v1/avatars/{avatar_id} [get]
func (h *Handler) AvatarGetHandler(w http.ResponseWriter, r *http.Request) {
	avatarID, ok := h.avatarIDParam(w, r)
	if !ok {
		return
	}
	size, ok := h.sizeParam(w, r)
	if !ok {
		return
	}

	avatar, content, err := h.service.GetAvatarContent(r.Context(), avatarID, size)
	if errors.Is(err, domain.ErrAvatarNotFound) {
		h.writeAvatarNotFound(w)
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	h.serveAvatarContent(w, r, avatar, size, content)
}

// UserAvatarGetHandler godoc
// @Summary  Download the latest avatar of a user
// @Tags     avatars
// @Produce  image/jpeg,image/png,image/webp
// @Param    user_id path int true "user id"
// @Param    size query string false "100x100, 300x300 or original"
// @Success  200
// @Failure  400 {object} handler.apiError
// @Failure  404 {object} handler.apiError
// @Failure  500
// @Router   /api/v1/users/{user_id}/avatar [get]
func (h *Handler) UserAvatarGetHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDParam(w, r)
	if !ok {
		return
	}
	size, ok := h.sizeParam(w, r)
	if !ok {
		return
	}

	avatar, content, err := h.service.GetUserAvatarContent(r.Context(), userID, size)
	if errors.Is(err, domain.ErrAvatarNotFound) {
		h.writeAvatarNotFound(w)
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	h.serveAvatarContent(w, r, avatar, size, content)
}
