package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/audit"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/repository"
)

// UserRegisterHandler godoc
// @Summary  Register a new user (returns a JWT)
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    request body domain.UserRegisterRequest true "login + password"
// @Success  200 {object} domain.UserRegisterResponse
// @Failure  400
// @Failure  409
// @Failure  500
// @Router   /api/user/register [post]
func (h *Handler) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "user.register")

	login, password, ok := h.decodeCredentials(w, r)
	if !ok {
		return
	}

	authResult, err := h.service.CreateUser(r.Context(),
		domain.UserRegisterRequest{Login: login, Password: password})
	if errors.Is(err, repository.ErrUserAlreadyExists) {
		h.writeAlreadyExists(w, err.Error())
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	h.writeToken(w, authResult.Token)
}
