package handler

import (
	"context"
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/logging"
	"go.uber.org/zap"
)

type Handler struct {
	service domain.GophprofileService
	logger  *zap.Logger
}

func NewHandler(service domain.GophprofileService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) writeUnauthorised(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("unauthorised", zap.String("msg", msg))
	http.Error(w, msg, http.StatusUnauthorized)
}

func (h *Handler) writeAlreadyExists(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("already exists", zap.String("msg", msg))
	http.Error(w, msg, http.StatusConflict)
}

func (h *Handler) writeInternalServerError(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Error("internal server error", zap.String("msg", msg))
	http.Error(w, msg, http.StatusInternalServerError)
}

func (h *Handler) writeBadRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("bad request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusBadRequest)
}

func (h *Handler) writeNotFound(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("not found request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusNotFound)
}

func (h *Handler) writeConflict(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("conflict", zap.String("msg", msg))
	http.Error(w, msg, http.StatusConflict)
}

func (h *Handler) writeMethodNotAllowed(ctx context.Context, w http.ResponseWriter, msg string) {
	logging.From(ctx).Info("method not allowed", zap.String("msg", msg))
	http.Error(w, msg, http.StatusMethodNotAllowed)
}
