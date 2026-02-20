package parser

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta represents the parsed SKILL.md frontmatter
type SkillMeta struct {
	Name            string   `yaml:"name"`
	Version         string   `yaml:"version"`
	Description     string   `yaml:"description"`
	Author          string   `yaml:"author"`
	Tags            []string `yaml:"tags"`
	License         string   `yaml:"license"`
	MinAgentVersion string   `yaml:"min_agent_version"`
	Compatibility   string   `yaml:"compatibility"`
}

var nameRegex = regexp.MustCompile(`^[a-z0-9\-]{3,64}$`)
var tagRegex = regexp.MustCompile(`^[a-z0-9\-]{1,32}$`)
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// ParseSkillMD parses and validates SKILL.md content
func ParseSkillMD(content string) (*SkillMeta, string, error) {
	// Split frontmatter from body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("SKILL.md must have YAML frontmatter delimited by ---")
	}

	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, "", fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	body := strings.TrimSpace(parts[2])

	// Validate required fields
	if err := validateMeta(&meta); err != nil {
		return nil, "", err
	}

	return &meta, body, nil
}

func validateMeta(m *SkillMeta) error {
	// Name
	if m.Name == "" {
		return fmt.Errorf("field 'name' is required")
	}
	if !nameRegex.MatchString(m.Name) {
		return fmt.Errorf("field 'name' must match [a-z0-9\\-]{3,64}")
	}
	if strings.Contains(m.Name, "--") {
		return fmt.Errorf("field 'name' must not contain consecutive hyphens '--'")
	}
	if strings.HasPrefix(m.Name, "-") || strings.HasSuffix(m.Name, "-") {
		return fmt.Errorf("field 'name' must not start or end with a hyphen")
	}

	// Version
	if m.Version == "" {
		return fmt.Errorf("field 'version' is required")
	}
	if !semverRegex.MatchString(m.Version) {
		return fmt.Errorf("field 'version' must be valid semver, got '%s'", m.Version)
	}

	// Description
	if m.Description == "" {
		return fmt.Errorf("field 'description' is required")
	}
	if len(m.Description) > 256 {
		return fmt.Errorf("field 'description' must be 1-256 characters, got %d", len(m.Description))
	}

	// Author
	if m.Author == "" {
		return fmt.Errorf("field 'author' is required")
	}

	// Tags (optional)
	if len(m.Tags) > 10 {
		return fmt.Errorf("field 'tags' allows max 10 items, got %d", len(m.Tags))
	}
	for _, tag := range m.Tags {
		if !tagRegex.MatchString(tag) {
			return fmt.Errorf("tag '%s' must match [a-z0-9\\-]{1,32}", tag)
		}
	}

	return nil
}
