//go:build server

package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/liuyukai/agentskills/internal/auth"
)

// Token format: ask_<40 hex chars> (44 chars total)
const tokenPrefix = "ask_"

// CreateToken creates a new Personal Access Token.
// POST /v1/tokens
// Body: {"name": "my-token", "expires_in_days": 90}
func (h *Handlers) CreateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name          string `json:"name"`
		ExpiresInDays *int   `json:"expires_in_days,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Token name is required")
		return
	}
	if len(req.Name) > 64 {
		writeError(w, http.StatusBadRequest, "Token name must be at most 64 characters")
		return
	}

	// Generate random token
	rawBytes := make([]byte, 20)
	if _, err := rand.Read(rawBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	plainToken := tokenPrefix + hex.EncodeToString(rawBytes)

	// Hash for storage
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Prefix for display (first 12 chars)
	displayPrefix := plainToken[:12] + "..."

	// Expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		t := time.Now().UTC().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &t
	}

	token, err := h.DB.CreateToken(ctx, user.ID, req.Name, tokenHash, displayPrefix, expiresAt)
	if err != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("Token name %q already exists", req.Name))
		return
	}

	// Return the plain token ONLY at creation time
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           token.ID,
		"name":         token.Name,
		"token":        plainToken,
		"token_prefix": displayPrefix,
		"expires_at":   token.ExpiresAt,
		"created_at":   token.CreatedAt,
		"message":      "Save this token — it won't be shown again.",
	})
}

// ListTokens lists all PATs for the authenticated user.
// GET /v1/tokens
func (h *Handlers) ListTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokens, err := h.DB.ListTokens(ctx, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	result := make([]map[string]interface{}, 0, len(tokens))
	for _, t := range tokens {
		item := map[string]interface{}{
			"id":           t.ID,
			"name":         t.Name,
			"token_prefix": t.TokenPrefix,
			"created_at":   t.CreatedAt,
		}
		if t.ExpiresAt != nil {
			item["expires_at"] = t.ExpiresAt
		}
		if t.LastUsedAt != nil {
			item["last_used_at"] = t.LastUsedAt
		}
		result = append(result, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tokens": result,
	})
}

// DeleteToken revokes a PAT.
// DELETE /v1/tokens/{id}
func (h *Handlers) DeleteToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenID := chi.URLParam(r, "id")
	if err := h.DB.DeleteToken(ctx, tokenID, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "Token not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Token revoked successfully",
	})
}
