package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_ClaudeDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	result := Detect(dir)
	if result.Provider != Claude {
		t.Errorf("expected Claude, got %s", result.Provider)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence, got %s", result.Confidence)
	}
}

func TestDetect_ClaudeMd(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude"), 0o644)

	result := Detect(dir)
	if result.Provider != Claude {
		t.Errorf("expected Claude, got %s", result.Provider)
	}
}

func TestDetect_GeminiDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".gemini"), 0o755)

	result := Detect(dir)
	if result.Provider != Gemini {
		t.Errorf("expected Gemini, got %s", result.Provider)
	}
}

func TestDetect_GeminiMd(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("# Gemini"), 0o644)

	result := Detect(dir)
	if result.Provider != Gemini {
		t.Errorf("expected Gemini, got %s", result.Provider)
	}
}

func TestDetect_CodexDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".codex"), 0o755)

	result := Detect(dir)
	if result.Provider != Codex {
		t.Errorf("expected Codex, got %s", result.Provider)
	}
}

func TestDetect_CopilotInstructions(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".github"), 0o755)
	os.WriteFile(filepath.Join(dir, ".github", "copilot-instructions.md"), []byte("# Copilot"), 0o644)

	result := Detect(dir)
	if result.Provider != Copilot {
		t.Errorf("expected Copilot, got %s", result.Provider)
	}
}

func TestDetect_CursorDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".cursor"), 0o755)

	result := Detect(dir)
	if result.Provider != Cursor {
		t.Errorf("expected Cursor, got %s", result.Provider)
	}
}

func TestDetect_CursorRules(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".cursorrules"), []byte("rules"), 0o644)

	result := Detect(dir)
	if result.Provider != Cursor {
		t.Errorf("expected Cursor, got %s", result.Provider)
	}
}

func TestDetect_WindsurfDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".windsurf"), 0o755)

	result := Detect(dir)
	if result.Provider != Windsurf {
		t.Errorf("expected Windsurf, got %s", result.Provider)
	}
}

func TestDetect_WindsurfRules(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".windsurfrules"), []byte("rules"), 0o644)

	result := Detect(dir)
	if result.Provider != Windsurf {
		t.Errorf("expected Windsurf, got %s", result.Provider)
	}
}

func TestDetect_AntigravityDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".antigravity"), 0o755)

	result := Detect(dir)
	if result.Provider != Antigravity {
		t.Errorf("expected Antigravity, got %s", result.Provider)
	}
}

func TestDetect_ClaudeSkillsDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0o755)

	result := Detect(dir)
	if result.Provider != Claude {
		t.Errorf("expected Claude, got %s", result.Provider)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence, got %s", result.Confidence)
	}
}

func TestDetect_NoIndicators(t *testing.T) {
	dir := t.TempDir()

	result := Detect(dir)
	if result.Provider != Generic {
		t.Errorf("expected Generic, got %s", result.Provider)
	}
	if result.Confidence != "low" {
		t.Errorf("expected low confidence, got %s", result.Confidence)
	}
}

func TestDetect_BothClaudeAndGemini(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".gemini"), 0o755)

	result := Detect(dir)
	// Both have score 3, should be ambiguous → generic
	if result.Provider != Generic {
		t.Errorf("expected Generic (ambiguous), got %s", result.Provider)
	}
}

func TestValidateName_ClaudeWithAnthropic(t *testing.T) {
	err := ValidateName(Claude, "my-anthropic-skill")
	if err == nil {
		t.Error("expected error for 'anthropic' in Claude skill name")
	}
}

func TestValidateName_ClaudeWithClaude(t *testing.T) {
	err := ValidateName(Claude, "claude-helper")
	if err == nil {
		t.Error("expected error for 'claude' in Claude skill name")
	}
}

func TestValidateName_ClaudeValid(t *testing.T) {
	err := ValidateName(Claude, "code-review")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateName_GenericAllowsClaude(t *testing.T) {
	err := ValidateName(Generic, "claude-helper")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsValidProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"claude", true},
		{"gemini", true},
		{"codex", true},
		{"copilot", true},
		{"cursor", true},
		{"windsurf", true},
		{"antigravity", true},
		{"generic", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsValidProvider(tt.input); got != tt.expected {
			t.Errorf("IsValidProvider(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestWorkspaceInstallPath(t *testing.T) {
	tests := []struct {
		provider Provider
		contains string
	}{
		{Claude, ".claude/skills/my-skill"},
		{Gemini, ".agents/skills/my-skill"},
		{Codex, ".agents/skills/my-skill"},
		{Copilot, ".github/skills/my-skill"},
		{Cursor, ".cursor/skills/my-skill"},
		{Windsurf, ".windsurf/skills/my-skill"},
		{Antigravity, ".agent/skills/my-skill"},
		{Generic, "my-skill"},
	}

	for _, tt := range tests {
		path := WorkspaceInstallPath(tt.provider, "my-skill", "/tmp/project")
		if !contains(path, tt.contains) {
			t.Errorf("WorkspaceInstallPath(%s) = %s, expected to contain %s", tt.provider, path, tt.contains)
		}
	}
}

func TestSkillTemplate(t *testing.T) {
	tmpl := SkillTemplate(Claude, "test-skill")
	if !contains(tmpl, "Claude Code") {
		t.Error("Claude template should mention Claude Code in compatibility")
	}
	if !contains(tmpl, "test-skill") {
		t.Error("Template should contain skill name")
	}

	tmplGeneric := SkillTemplate(Generic, "test-skill")
	if contains(tmplGeneric, "compatibility") {
		t.Error("Generic template should not have compatibility field")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
