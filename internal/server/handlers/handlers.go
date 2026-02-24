//go:build server

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/liuyukai/agentskills/internal/database"
	"github.com/liuyukai/agentskills/internal/storage"
)

type Handlers struct {
	DB    database.Database
	Store storage.Storage
}

func New(db database.Database, store storage.Storage) *Handlers {
	return &Handlers{DB: db, Store: store}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
