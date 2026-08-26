package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
)

// AvatarMetadataHandler godoc
// @Summary  Get avatar metadata
// @Tags     avatars
// @Produce  json
// @Param    avatar_id path string true "avatar UUID"
// @Success  200 {object} domain.AvatarMetadataResponse
// @Failure  404 {object} handler.apiError
// @Failure  500
// @Router   /api/v1/avatars/{avatar_id}/metadata [get]
func (h *Handler) AvatarMetadataHandler(w http.ResponseWriter, r *http.Request) {
	avatarID, ok := h.avatarIDParam(w, r)
	if !ok {
		return
	}

	avatar, err := h.service.GetAvatarMetadata(r.Context(), avatarID)
	if errors.Is(err, domain.ErrAvatarNotFound) {
		h.writeAvatarNotFound(w)
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, avatar.ToMetadataResponse())
}

// UserAvatarsListHandler godoc
// @Summary  List all avatars of a user
// @Tags     avatars
// @Produce  json
// @Param    user_id path int true "user id"
// @Success  200 {object} domain.AvatarListResponse
// @Failure  400 {object} handler.apiError
// @Failure  500
// @Router   /api/v1/users/{user_id}/avatars [get]
func (h *Handler) UserAvatarsListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDParam(w, r)
	if !ok {
		return
	}

	avatars, err := h.service.ListUserAvatars(r.Context(), userID)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	response := domain.AvatarListResponse{
		Avatars: make([]domain.AvatarMetadataResponse, 0, len(avatars)),
	}
	for i := range avatars {
		response.Avatars = append(response.Avatars, avatars[i].ToMetadataResponse())
	}

	h.writeJSON(w, http.StatusOK, response)
}
