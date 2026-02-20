package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider represents an AI agent provider
type Provider string

const (
	Claude       Provider = "claude"
	Gemini       Provider = "gemini"
	Codex        Provider = "codex"
	Copilot      Provider = "copilot"
	Cursor       Provider = "cursor"
	Windsurf     Provider = "windsurf"
	Antigravity  Provider = "antigravity"
	Generic      Provider = "generic"
)

// AllProviders lists all known providers (excluding generic)
var AllProviders = []Provider{Claude, Gemini, Codex, Copilot, Cursor, Windsurf, Antigravity}

// DetectResult holds the result of auto-detection
type DetectResult struct {
	Provider   Provider
	Confidence string   // "explicit", "high", "low"
	Indicators []string // what was found
}

type indicator struct {
	path     string
	provider Provider
	score    int
	isDir    bool
}

var indicators = []indicator{
	// Claude
	{".claude", Claude, 3, true},
	{"CLAUDE.md", Claude, 2, false},
	{".claude/skills", Claude, 4, true},

	// Gemini
	{".gemini", Gemini, 3, true},
	{"GEMINI.md", Gemini, 2, false},
	{".gemini/skills", Gemini, 4, true},
	{".agents", Gemini, 1, true},
	{".agents/skills", Gemini, 3, true},

	// Codex
	{".codex", Codex, 4, true},
	{"AGENTS.md", Codex, 2, false},

	// Copilot
	{".github/copilot-instructions.md", Copilot, 4, false},
	{".github/skills", Copilot, 4, true},
	{".github/agents", Copilot, 3, true},

	// Cursor
	{".cursor", Cursor, 3, true},
	{".cursorrules", Cursor, 3, false},
	{".cursor/rules", Cursor, 4, true},

	// Windsurf
	{".windsurf", Windsurf, 3, true},
	{".windsurfrules", Windsurf, 3, false},
	{".windsurf/rules", Windsurf, 4, true},

	// Antigravity
	{".antigravity", Antigravity, 4, true},
	{".antigravity/rules.md", Antigravity, 3, false},
}

// Detect examines the given directory for provider indicators
func Detect(dir string) DetectResult {
	scores := make(map[Provider]int)
	var found []string

	for _, ind := range indicators {
		fullPath := filepath.Join(dir, ind.path)
		var exists bool
		if ind.isDir {
			info, err := os.Stat(fullPath)
			exists = err == nil && info.IsDir()
		} else {
			info, err := os.Stat(fullPath)
			exists = err == nil && !info.IsDir()
		}

		if exists {
			scores[ind.provider] += ind.score
			found = append(found, fmt.Sprintf("%s found", ind.path))
		}
	}

	// Find the winner
	var best Provider
	var bestScore int
	var tie bool

	for _, p := range AllProviders {
		s := scores[p]
		if s > bestScore {
			bestScore = s
			best = p
			tie = false
		} else if s == bestScore && s > 0 {
			tie = true
		}
	}

	if bestScore < 2 || tie {
		return DetectResult{
			Provider:   Generic,
			Confidence: "low",
			Indicators: found,
		}
	}

	return DetectResult{
		Provider:   best,
		Confidence: "high",
		Indicators: found,
	}
}

// WorkspaceInstallPath returns the project-level install path for a skill
func WorkspaceInstallPath(p Provider, skillName, cwd string) string {
	switch p {
	case Claude:
		return filepath.Join(cwd, ".claude", "skills", skillName)
	case Gemini:
		return filepath.Join(cwd, ".agents", "skills", skillName)
	case Codex:
		return filepath.Join(cwd, ".agents", "skills", skillName)
	case Copilot:
		return filepath.Join(cwd, ".github", "skills", skillName)
	case Cursor:
		return filepath.Join(cwd, ".cursor", "skills", skillName)
	case Windsurf:
		return filepath.Join(cwd, ".windsurf", "skills", skillName)
	case Antigravity:
		return filepath.Join(cwd, ".agent", "skills", skillName)
	default:
		return filepath.Join(cwd, skillName)
	}
}

// UserInstallPath returns the user-level install path for a skill
func UserInstallPath(p Provider, skillName string) string {
	home, _ := os.UserHomeDir()
	switch p {
	case Claude:
		return filepath.Join(home, ".claude", "skills", skillName)
	case Gemini:
		return filepath.Join(home, ".agents", "skills", skillName)
	case Codex:
		return filepath.Join(home, ".codex", "skills", skillName)
	case Copilot:
		return filepath.Join(home, skillName)
	case Cursor:
		return filepath.Join(home, ".cursor", "skills", skillName)
	case Windsurf:
		return filepath.Join(home, ".codeium", "skills", skillName)
	case Antigravity:
		return filepath.Join(home, ".antigravity", "skills", skillName)
	default:
		return filepath.Join(".", skillName)
	}
}

// ValidateName applies provider-specific name restrictions
func ValidateName(p Provider, name string) error {
	if p == Claude {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "anthropic") || strings.Contains(lower, "claude") {
			return fmt.Errorf("skill name '%s' cannot contain 'anthropic' or 'claude' for Claude-compatible skills", name)
		}
	}
	return nil
}

// IsValidProvider checks if a string is a recognized provider
func IsValidProvider(s string) bool {
	switch Provider(s) {
	case Claude, Gemini, Codex, Copilot, Cursor, Windsurf, Antigravity, Generic:
		return true
	}
	return false
}

// SkillTemplate returns the SKILL.md template content for a provider
func SkillTemplate(p Provider, skillName string) string {
	compat := compatibilityValue(p)
	compatLine := ""
	if compat != "" {
		compatLine = fmt.Sprintf("compatibility: %q\n", compat)
	}

	return fmt.Sprintf(`---
name: %q
version: "0.1.0"
description: %q
author: ""
tags: []
%s---

# %s

## When to use this skill
Use this skill when...

## Instructions
1. Step one...
2. Step two...

## Examples
[Concrete examples of using this skill]
`, skillName, descriptionHint(p), compatLine, skillName)
}

func compatibilityValue(p Provider) string {
	switch p {
	case Claude:
		return "Designed for Claude Code"
	case Gemini:
		return "Designed for Gemini CLI"
	case Codex:
		return "Designed for OpenAI Codex"
	case Copilot:
		return "Designed for VS Code Copilot"
	case Cursor:
		return "Designed for Cursor IDE"
	case Windsurf:
		return "Designed for Windsurf"
	case Antigravity:
		return "Designed for Antigravity"
	default:
		return ""
	}
}

func descriptionHint(p Provider) string {
	switch p {
	case Claude:
		return "Brief description of what this skill does and when Claude should use it."
	default:
		return "Brief description of what this skill does and when to use it."
	}
}
