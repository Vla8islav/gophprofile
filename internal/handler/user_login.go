package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/audit"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/repository"
)

// UserLoginHandler godoc
// @Summary  Authenticate (returns a JWT)
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    request body domain.UserLoginRequest true "login + password"
// @Success  200 {object} domain.UserLoginResponse
// @Failure  400
// @Failure  401
// @Failure  500
// @Router   /api/user/login [post]
func (h *Handler) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "user.login")
	login, password, ok := h.decodeCredentials(w, r)
	if !ok {
		return
	}
	authResult, err := h.service.LoginUser(r.Context(),
		domain.UserLoginRequest{Login: login, Password: password})
	if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, domain.ErrInvalidUserCredentials) {
		h.writeUnauthorised(w, "invalid user login or password")
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}
	h.writeToken(w, authResult.Token)
}
