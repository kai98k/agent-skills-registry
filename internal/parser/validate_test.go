package parser

import (
	"testing"
)

func TestValidate_Valid(t *testing.T) {
	meta := &SkillMeta{
		Name:        "code-review-agent",
		Version:     "1.0.0",
		Description: "A code review skill",
		Author:      "dev",
		Tags:        []string{"code-review", "github"},
	}
	if err := Validate(meta); err != nil {
		t.Fatalf("expected valid, got error: %v", err)
	}
}

func TestValidate_MissingName(t *testing.T) {
	meta := &SkillMeta{Version: "1.0.0", Description: "test", Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidate_ShortName(t *testing.T) {
	meta := &SkillMeta{Name: "ab", Version: "1.0.0", Description: "test", Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for short name")
	}
}

func TestValidate_UppercaseName(t *testing.T) {
	meta := &SkillMeta{Name: "Code-Review", Version: "1.0.0", Description: "test", Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for uppercase name")
	}
}

func TestValidate_ConsecutiveHyphens(t *testing.T) {
	meta := &SkillMeta{Name: "code--review", Version: "1.0.0", Description: "test", Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for consecutive hyphens")
	}
}

func TestValidate_InvalidSemver(t *testing.T) {
	cases := []string{"1.0", "v1.0.0", "1.0.0.0", "abc", "1.0.0-beta", "01.0.0"}
	for _, v := range cases {
		meta := &SkillMeta{Name: "test-skill", Version: v, Description: "test", Author: "dev"}
		if err := Validate(meta); err == nil {
			t.Fatalf("expected error for version %q", v)
		}
	}
}

func TestValidate_ValidSemver(t *testing.T) {
	cases := []string{"0.1.0", "1.0.0", "10.20.30", "0.0.1"}
	for _, v := range cases {
		meta := &SkillMeta{Name: "test-skill", Version: v, Description: "test", Author: "dev"}
		if err := Validate(meta); err != nil {
			t.Fatalf("expected valid for version %q, got: %v", v, err)
		}
	}
}

func TestValidate_MissingDescription(t *testing.T) {
	meta := &SkillMeta{Name: "test-skill", Version: "1.0.0", Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestValidate_LongDescription(t *testing.T) {
	desc := ""
	for i := 0; i < 257; i++ {
		desc += "a"
	}
	meta := &SkillMeta{Name: "test-skill", Version: "1.0.0", Description: desc, Author: "dev"}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for long description")
	}
}

func TestValidate_TooManyTags(t *testing.T) {
	tags := make([]string, 11)
	for i := range tags {
		tags[i] = "tag"
	}
	meta := &SkillMeta{Name: "test-skill", Version: "1.0.0", Description: "test", Author: "dev", Tags: tags}
	if err := Validate(meta); err == nil {
		t.Fatal("expected error for too many tags")
	}
}
