package handler

import (
	"context"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func newAvatarTestHandler(service domain.GophprofileService) *Handler {
	return &Handler{
		service: service,
		logger:  zap.NewNop(),
	}
}

// withChiParams injects chi URL params to avoid spinning up the full router.
func withChiParams(r *http.Request, params map[string]string) *http.Request {
	routeCtx := chi.NewRouteContext()
	for key, value := range params {
		routeCtx.URLParams.Add(key, value)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

// asUser simulates the WithAuth middleware having authenticated the request.
func asUser(r *http.Request, userID int64) *http.Request {
	return r.WithContext(middlewares.ContextWithUserID(r.Context(), userID))
}
