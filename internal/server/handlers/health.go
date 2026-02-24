//go:build server

package handlers

import "net/http"

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbStatus := "connected"
	if err := h.DB.HealthCheck(ctx); err != nil {
		dbStatus = "disconnected"
	}

	storageStatus := "connected"
	if err := h.Store.HealthCheck(ctx); err != nil {
		storageStatus = "disconnected"
	}

	status := http.StatusOK
	overall := "ok"
	if dbStatus != "connected" || storageStatus != "connected" {
		status = http.StatusServiceUnavailable
		overall = "degraded"
	}

	writeJSON(w, status, map[string]string{
		"status":   overall,
		"database": dbStatus,
		"storage":  storageStatus,
	})
}
