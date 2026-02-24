//go:build server

package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) GetSkill(w http.ResponseWriter, r *http.Request) {
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

	owner, err := h.DB.GetUserByID(ctx, skill.OwnerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	ownerName := ""
	if owner != nil {
		ownerName = owner.Username
	}

	latest, err := h.DB.GetLatestVersion(ctx, skill.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	resp := map[string]interface{}{
		"name":       skill.Name,
		"owner":      ownerName,
		"downloads":  skill.Downloads,
		"created_at": skill.CreatedAt,
	}

	if latest != nil {
		resp["latest_version"] = map[string]interface{}{
			"version":      latest.Version,
			"description":  metaStr(latest.Metadata, "description"),
			"checksum":     "sha256:" + latest.Checksum,
			"size_bytes":   latest.SizeBytes,
			"published_at": latest.PublishedAt,
			"metadata":     latest.Metadata,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func metaStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}
