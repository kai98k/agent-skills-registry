//go:build server

package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/liuyukai/agentskills/internal/bundle"
	"github.com/liuyukai/agentskills/internal/database"
	"github.com/liuyukai/agentskills/internal/parser"
	"github.com/liuyukai/agentskills/internal/auth"
)

const maxBundleSize = 50 * 1024 * 1024 // 50MB

func (h *Handlers) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Limit request body
	r.Body = http.MaxBytesReader(w, r.Body, maxBundleSize)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing or invalid file upload")
		return
	}
	defer file.Close()

	// Read file into memory for checksum + processing
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		if err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "Bundle exceeds 50MB limit")
			return
		}
		writeError(w, http.StatusBadRequest, "Failed to read uploaded file")
		return
	}
	fileBytes := buf.Bytes()

	// Compute checksum
	checksum := bundle.SHA256Bytes(fileBytes)

	// Unpack to temp dir
	tmpDir, err := os.MkdirTemp("", "agentskills-publish-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create temp directory")
		return
	}
	defer os.RemoveAll(tmpDir)

	if err := bundle.UnpackReader(bytes.NewReader(fileBytes), tmpDir); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to unpack archive: %v", err))
		return
	}

	// Find and parse SKILL.md
	skillMDPath, err := bundle.FindSKILLMD(tmpDir)
	if err != nil {
		writeError(w, http.StatusBadRequest, "No SKILL.md found in archive")
		return
	}

	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to read SKILL.md")
		return
	}

	meta, _, err := parser.ParseSKILLMD(content)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid SKILL.md: %v", err))
		return
	}

	if err := parser.Validate(meta); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	// Author must match authenticated user
	if meta.Author != user.Username {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("author %q does not match authenticated user %q", meta.Author, user.Username))
		return
	}

	// Check skill ownership
	skill, err := h.DB.GetSkillByName(ctx, meta.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if skill == nil {
		// Create new skill
		skill, err = h.DB.CreateSkill(ctx, meta.Name, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create skill")
			return
		}
	} else if skill.OwnerID != user.ID {
		writeError(w, http.StatusForbidden, fmt.Sprintf("Skill '%s' is owned by another user", meta.Name))
		return
	}

	// Check version doesn't exist
	existing, err := h.DB.GetVersion(ctx, skill.ID, meta.Version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("Version %s already exists", meta.Version))
		return
	}

	// Upload to storage
	bundleKey := fmt.Sprintf("%s/%s.tar.gz", meta.Name, meta.Version)
	if err := h.Store.Upload(ctx, bundleKey, bytes.NewReader(fileBytes), int64(len(fileBytes))); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to store bundle")
		return
	}

	// Create version record
	sv := &database.SkillVersion{
		SkillID:   skill.ID,
		Version:   meta.Version,
		BundleKey: bundleKey,
		Metadata:  meta.ToMap(),
		Checksum:  checksum,
		SizeBytes: int64(len(fileBytes)),
	}
	if err := h.DB.CreateVersion(ctx, sv); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save version")
		return
	}

	// Update skill timestamp
	h.DB.UpdateSkillTimestamp(ctx, skill.ID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"name":         meta.Name,
		"version":      meta.Version,
		"checksum":     "sha256:" + checksum,
		"published_at": sv.PublishedAt,
	})
}
