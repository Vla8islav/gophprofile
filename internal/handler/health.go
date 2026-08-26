package handler

import (
	"context"
	"net/http"
	"time"
)

type healthResponse struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
}

const healthCheckTimeout = 3 * time.Second

// HealthHandler godoc
// @Summary  Health check with per-component statuses
// @Tags     system
// @Produce  json
// @Success  200 {object} handler.healthResponse
// @Failure  503 {object} handler.healthResponse
// @Router   /health [get]
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
	defer cancel()

	response := healthResponse{
		Status: "ok",
		Components: map[string]string{
			"database": "ok",
			"s3":       "ok",
			"broker":   "ok",
		},
	}

	if err := h.service.Ping(ctx); err != nil {
		response.Components["database"] = "error: " + err.Error()
		response.Status = "degraded"
	}
	if err := h.service.FileStoragePing(ctx); err != nil {
		response.Components["s3"] = "error: " + err.Error()
		response.Status = "degraded"
	}
	if err := h.service.BrokerPing(ctx); err != nil {
		response.Components["broker"] = "error: " + err.Error()
		response.Status = "degraded"
	}

	status := http.StatusOK
	if response.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	h.writeJSON(w, status, response)
}
