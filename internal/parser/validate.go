package parser

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,62}[a-z0-9]$`)
	tagRegex  = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{0,30}[a-z0-9]$`)
)

// Validate checks all frontmatter fields according to SDD §2.3.
func Validate(meta *SkillMeta) error {
	// name: required, [a-z0-9-], 3-64 chars, no consecutive --
	if meta.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(meta.Name) < 3 || len(meta.Name) > 64 {
		return fmt.Errorf("name must be 3-64 characters, got %d", len(meta.Name))
	}
	if !nameRegex.MatchString(meta.Name) {
		return fmt.Errorf("name must match [a-z0-9-], got %q", meta.Name)
	}
	if strings.Contains(meta.Name, "--") {
		return fmt.Errorf("name must not contain consecutive hyphens (--)")
	}

	// version: required, strict semver
	if meta.Version == "" {
		return fmt.Errorf("version is required")
	}
	if !isValidSemver(meta.Version) {
		return fmt.Errorf("version must be strict semver (MAJOR.MINOR.PATCH), got %q", meta.Version)
	}

	// description: required, 1-256 chars
	if meta.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(meta.Description) > 256 {
		return fmt.Errorf("description must be at most 256 characters, got %d", len(meta.Description))
	}

	// author: required
	if meta.Author == "" {
		return fmt.Errorf("author is required")
	}

	// tags: optional, max 10, each [a-z0-9-] 1-32 chars
	if len(meta.Tags) > 10 {
		return fmt.Errorf("at most 10 tags allowed, got %d", len(meta.Tags))
	}
	for _, tag := range meta.Tags {
		if len(tag) < 1 || len(tag) > 32 {
			return fmt.Errorf("tag must be 1-32 characters, got %q (%d chars)", tag, len(tag))
		}
		// Single char tags are OK
		if len(tag) == 1 {
			if !regexp.MustCompile(`^[a-z0-9]$`).MatchString(tag) {
				return fmt.Errorf("tag must match [a-z0-9-], got %q", tag)
			}
		} else if !tagRegex.MatchString(tag) {
			return fmt.Errorf("tag must match [a-z0-9-], got %q", tag)
		}
	}

	return nil
}

// isValidSemver checks for strict MAJOR.MINOR.PATCH format.
func isValidSemver(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		// Must be numeric
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
		// No leading zeros (except "0" itself)
		if len(p) > 1 && p[0] == '0' {
			return false
		}
	}
	return true
}
