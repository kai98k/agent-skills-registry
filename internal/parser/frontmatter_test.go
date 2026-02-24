package parser

import (
	"testing"
)

func TestParseSKILLMD_Valid(t *testing.T) {
	content := []byte(`---
name: "test-skill"
version: "0.1.0"
description: "A test skill"
author: "dev"
tags:
  - test
---

# Test Skill

Instructions here.
`)

	meta, body, err := ParseSKILLMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "test-skill" {
		t.Errorf("name = %q, want %q", meta.Name, "test-skill")
	}
	if meta.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", meta.Version, "0.1.0")
	}
	if meta.Description != "A test skill" {
		t.Errorf("description = %q, want %q", meta.Description, "A test skill")
	}
	if meta.Author != "dev" {
		t.Errorf("author = %q, want %q", meta.Author, "dev")
	}
	if len(meta.Tags) != 1 || meta.Tags[0] != "test" {
		t.Errorf("tags = %v, want [test]", meta.Tags)
	}
	if body == "" {
		t.Error("body should not be empty")
	}
}

func TestParseSKILLMD_NoFrontmatter(t *testing.T) {
	content := []byte("# Just a markdown file\n\nNo frontmatter here.")
	_, _, err := ParseSKILLMD(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParseSKILLMD_UnclosedFrontmatter(t *testing.T) {
	content := []byte("---\nname: test\n# No closing ---")
	_, _, err := ParseSKILLMD(content)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
}
