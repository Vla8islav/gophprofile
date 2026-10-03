package handler

import (
	"net/http"

	"github.com/Vla8islav/gophprofile/internal/logging"
	"go.uber.org/zap"
)

// DBPing godoc
// @Summary  Health check (database ping)
// @Tags     system
// @Success  200
// @Failure  500
// @Router   /api/ping [get]
func (h *Handler) DBPing(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		logging.From(r.Context()).Warn("method not allowed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.service.Ping(r.Context())
	if err != nil {
		logging.From(r.Context()).Error("db ping failed",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
