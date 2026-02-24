//go:build server

package handlers

import (
	"net/http"
	"strconv"

	"github.com/liuyukai/agentskills/internal/database"
)

func (h *Handlers) SearchSkills(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	query := q.Get("q")
	tag := q.Get("tag")

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	results, total, err := h.DB.SearchSkills(ctx, query, tag, page, perPage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if results == nil {
		results = []database.SkillSearchResult{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"results":  results,
	})
}
