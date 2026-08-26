package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/audit"
	"github.com/Vla8islav/gophprofile/internal/domain"
)

func (h *Handler) writeDeleteResult(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAvatarNotFound):
		h.writeAvatarNotFound(w)
	case errors.Is(err, domain.ErrNotAvatarOwner):
		h.writeJSONError(w, http.StatusForbidden, apiError{
			Error:   "Forbidden",
			Details: "You can only delete your own avatars",
		})
	case err != nil:
		h.writeInternalServerError(w, err.Error())
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// AvatarDeleteHandler godoc
// @Summary   Delete an avatar (soft delete, own avatars only)
// @Tags      avatars
// @Param     avatar_id path string true "avatar UUID"
// @Success   204
// @Failure   401
// @Failure   403 {object} handler.apiError
// @Failure   404 {object} handler.apiError
// @Failure   500
// @Security  BearerAuth
// @Router    /api/v1/avatars/{avatar_id} [delete]
func (h *Handler) AvatarDeleteHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "avatar.delete")

	requesterID, ok := h.requestUserID(w, r)
	if !ok {
		return
	}
	avatarID, ok := h.avatarIDParam(w, r)
	if !ok {
		return
	}

	h.writeDeleteResult(w, h.service.DeleteAvatar(r.Context(), avatarID, requesterID))
}

// UserAvatarDeleteHandler godoc
// @Summary   Delete the latest avatar of a user (own avatars only)
// @Tags      avatars
// @Param     user_id path int true "user id"
// @Success   204
// @Failure   400 {object} handler.apiError
// @Failure   401
// @Failure   403 {object} handler.apiError
// @Failure   404 {object} handler.apiError
// @Failure   500
// @Security  BearerAuth
// @Router    /api/v1/users/{user_id}/avatar [delete]
func (h *Handler) UserAvatarDeleteHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "avatar.delete_user_avatar")

	requesterID, ok := h.requestUserID(w, r)
	if !ok {
		return
	}
	userID, ok := h.userIDParam(w, r)
	if !ok {
		return
	}

	h.writeDeleteResult(w, h.service.DeleteUserAvatar(r.Context(), userID, requesterID))
}
