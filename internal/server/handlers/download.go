//go:build server

package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")

	skill, err := h.DB.GetSkillByName(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if skill == nil {
		writeError(w, http.StatusNotFound, "Skill not found")
		return
	}

	sv, err := h.DB.GetVersion(ctx, skill.ID, version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if sv == nil {
		writeError(w, http.StatusNotFound, "Version not found")
		return
	}

	reader, size, err := h.Store.Download(ctx, sv.BundleKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}
	defer reader.Close()

	// Increment download count
	go h.DB.IncrementDownloads(r.Context(), skill.ID)

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.tar.gz"`, name, version))
	w.Header().Set("X-Checksum-SHA256", sv.Checksum)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.WriteHeader(http.StatusOK)

	io.Copy(w, reader)
}
