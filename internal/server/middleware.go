//go:build server

package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/liuyukai/agentskills/internal/auth"
	"github.com/liuyukai/agentskills/internal/database"
)

// AuthMiddleware extracts the Bearer token and resolves the user.
// It checks both the legacy users.api_token and the PAT tokens table.
func AuthMiddleware(db database.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			// 1. Try legacy api_token in users table
			user, err := db.GetUserByToken(ctx, token)
			if err != nil {
				http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
				return
			}

			// 2. Try PAT tokens table (hash the token first)
			if user == nil {
				hash := sha256.Sum256([]byte(token))
				tokenHash := hex.EncodeToString(hash[:])
				user, err = db.GetUserByTokenHash(ctx, tokenHash)
				if err != nil {
					http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
					return
				}
				// Update last_used_at in background
				if user != nil {
					go db.UpdateTokenLastUsed(context.Background(), tokenHash)
				}
			}

			if user == nil {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx = auth.WithUser(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
