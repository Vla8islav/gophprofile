package handler

import (
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	service domain.GophprofileService
	logger  *zap.Logger
}

func NewHandler(service domain.GophprofileService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) writeUnauthorised(w http.ResponseWriter, msg string) {
	h.logger.Info("unauthorised", zap.String("msg", msg))
	http.Error(w, msg, http.StatusUnauthorized)
}

func (h *Handler) writeAlreadyExists(w http.ResponseWriter, msg string) {
	h.logger.Info("already exists", zap.String("msg", msg))
	http.Error(w, msg, http.StatusConflict)
}

func (h *Handler) writeInternalServerError(w http.ResponseWriter, msg string) {
	h.logger.Error("internal server error", zap.String("msg", msg))
	http.Error(w, msg, http.StatusInternalServerError)
}

func (h *Handler) writeBadRequest(w http.ResponseWriter, msg string) {
	h.logger.Info("bad request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusBadRequest)
}

func (h *Handler) writeNotFound(w http.ResponseWriter, msg string) {
	h.logger.Info("not found request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusNotFound)
}

func (h *Handler) writeConflict(w http.ResponseWriter, msg string) {
	h.logger.Info("conflict", zap.String("msg", msg))
	http.Error(w, msg, http.StatusConflict)
}

func (h *Handler) writeMethodNotAllowed(w http.ResponseWriter, msg string) {
	h.logger.Info("method not allowed", zap.String("msg", msg))
	http.Error(w, msg, http.StatusMethodNotAllowed)
}
