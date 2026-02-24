//go:build server

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/liuyukai/agentskills/internal/bundle"
	"github.com/liuyukai/agentskills/internal/database"
	"github.com/liuyukai/agentskills/internal/server"
	"github.com/liuyukai/agentskills/internal/storage"
)

func setupTestServer(t *testing.T) (*server.Server, func()) {
	t.Helper()

	// Use temp file for SQLite to avoid shared :memory: issues
	tmpDir, err := os.MkdirTemp("", "agentskills-test-*")
	if err != nil {
		t.Fatalf("tmpdir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db := database.NewSQLite(dbPath)
	if err := db.Open(); err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Temp storage
	storePath := filepath.Join(tmpDir, "bundles")
	store := storage.NewLocalStorage(storePath)
	if err := store.Init(); err != nil {
		t.Fatalf("init storage: %v", err)
	}

	srv := server.New(db, store, 0)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return srv, cleanup
}

func createTestBundle(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agentskills-bundle-*")
	if err != nil {
		t.Fatal(err)
	}

	skillMD := `---
name: "test-skill"
version: "0.1.0"
description: "A test skill"
author: "dev"
tags:
  - test
---

# Test Skill

Test instructions.
`
	if err := os.WriteFile(tmpDir+"/SKILL.md", []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := tmpDir + ".tar.gz"
	if err := bundle.Pack(tmpDir, archivePath); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tmpDir)

	return archivePath
}

func TestHealthEndpoint(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestPublishAndGet(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	archivePath := createTestBundle(t)
	defer os.Remove(archivePath)

	// Publish
	body, contentType := createMultipartFile(t, archivePath)
	req := httptest.NewRequest("POST", "/v1/skills/publish", body)
	req.Header.Set("Authorization", "Bearer dev-token-12345")
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("publish status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var pubResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&pubResp)
	if pubResp["name"] != "test-skill" {
		t.Errorf("name = %v, want test-skill", pubResp["name"])
	}

	// Get skill
	req = httptest.NewRequest("GET", "/v1/skills/test-skill", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", w.Code, http.StatusOK)
	}

	var getResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&getResp)
	if getResp["name"] != "test-skill" {
		t.Errorf("name = %v, want test-skill", getResp["name"])
	}
}

func TestPublishNoAuth(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	archivePath := createTestBundle(t)
	defer os.Remove(archivePath)

	body, contentType := createMultipartFile(t, archivePath)
	req := httptest.NewRequest("POST", "/v1/skills/publish", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPublishDuplicateVersion(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	archivePath := createTestBundle(t)
	defer os.Remove(archivePath)

	// First publish
	body, contentType := createMultipartFile(t, archivePath)
	req := httptest.NewRequest("POST", "/v1/skills/publish", body)
	req.Header.Set("Authorization", "Bearer dev-token-12345")
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first publish status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Second publish (same version) → 409
	body, contentType = createMultipartFile(t, archivePath)
	req = httptest.NewRequest("POST", "/v1/skills/publish", body)
	req.Header.Set("Authorization", "Bearer dev-token-12345")
	req.Header.Set("Content-Type", contentType)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate publish status = %d, want %d, body: %s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestGetSkillNotFound(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/skills/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSearchEmpty(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/skills?q=nonexistent", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	results, _ := resp["results"].([]interface{})
	if len(results) != 0 {
		t.Errorf("results = %d, want 0", len(results))
	}
}

func TestTokenCreateAndUse(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Create token
	tokenReq := `{"name": "test-token", "expires_in_days": 30}`
	req := httptest.NewRequest("POST", "/v1/tokens", bytes.NewBufferString(tokenReq))
	req.Header.Set("Authorization", "Bearer dev-token-12345")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create token status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var tokenResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tokenResp)
	plainToken, ok := tokenResp["token"].(string)
	if !ok || plainToken == "" {
		t.Fatal("expected token in response")
	}

	// Use new PAT to access protected endpoint (list tokens)
	req = httptest.NewRequest("GET", "/v1/tokens", nil)
	req.Header.Set("Authorization", "Bearer "+plainToken)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list tokens with PAT status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// List tokens
	var listResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&listResp)
	tokens, _ := listResp["tokens"].([]interface{})
	if len(tokens) != 1 {
		t.Errorf("tokens count = %d, want 1", len(tokens))
	}
}

func createMultipartFile(t *testing.T, filePath string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	part, err := w.CreateFormFile("file", "bundle.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, f); err != nil {
		t.Fatal(err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}
