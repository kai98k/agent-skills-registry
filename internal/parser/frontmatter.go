package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta represents the parsed YAML frontmatter from SKILL.md.
type SkillMeta struct {
	Name            string   `yaml:"name" json:"name"`
	Version         string   `yaml:"version" json:"version"`
	Description     string   `yaml:"description" json:"description"`
	Author          string   `yaml:"author" json:"author"`
	Tags            []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	License         string   `yaml:"license,omitempty" json:"license,omitempty"`
	MinAgentVersion string   `yaml:"min_agent_version,omitempty" json:"min_agent_version,omitempty"`
}

// ParseSKILLMD parses the frontmatter and body from a SKILL.md file.
func ParseSKILLMD(content []byte) (*SkillMeta, string, error) {
	s := string(content)

	// Must start with ---
	if !strings.HasPrefix(strings.TrimSpace(s), "---") {
		return nil, "", fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}

	// Find the closing ---
	trimmed := strings.TrimSpace(s)
	rest := trimmed[3:] // skip opening ---
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, "", fmt.Errorf("SKILL.md frontmatter not closed (missing closing ---)")
	}

	frontmatterStr := rest[:idx]
	body := strings.TrimSpace(rest[idx+4:]) // skip \n---

	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(frontmatterStr), &meta); err != nil {
		return nil, "", fmt.Errorf("parse frontmatter YAML: %w", err)
	}

	return &meta, body, nil
}

// ToMap converts SkillMeta to a map for JSON storage.
func (m *SkillMeta) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"name":        m.Name,
		"version":     m.Version,
		"description": m.Description,
		"author":      m.Author,
	}
	if len(m.Tags) > 0 {
		tags := make([]interface{}, len(m.Tags))
		for i, t := range m.Tags {
			tags[i] = t
		}
		result["tags"] = tags
	}
	if m.License != "" {
		result["license"] = m.License
	}
	if m.MinAgentVersion != "" {
		result["min_agent_version"] = m.MinAgentVersion
	}
	return result
}
