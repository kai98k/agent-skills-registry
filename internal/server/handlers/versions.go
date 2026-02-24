//go:build server

package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) ListVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := chi.URLParam(r, "name")

	skill, err := h.DB.GetSkillByName(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if skill == nil {
		writeError(w, http.StatusNotFound, "Skill not found")
		return
	}

	versions, err := h.DB.ListVersions(ctx, skill.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	versionList := make([]map[string]interface{}, 0, len(versions))
	for _, v := range versions {
		versionList = append(versionList, map[string]interface{}{
			"version":      v.Version,
			"checksum":     "sha256:" + v.Checksum,
			"size_bytes":   v.SizeBytes,
			"published_at": v.PublishedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":     skill.Name,
		"versions": versionList,
	})
}
