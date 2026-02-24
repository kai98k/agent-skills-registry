package database

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	APIToken  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Token struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	TokenHash   string     `json:"-"`
	TokenPrefix string     `json:"token_prefix"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Skill struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	Downloads int64     `json:"downloads"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SkillVersion struct {
	ID          string                 `json:"id"`
	SkillID     string                 `json:"skill_id"`
	Version     string                 `json:"version"`
	BundleKey   string                 `json:"bundle_key"`
	Metadata    map[string]interface{} `json:"metadata"`
	Checksum    string                 `json:"checksum"`
	SizeBytes   int64                  `json:"size_bytes"`
	PublishedAt time.Time              `json:"published_at"`
}

type SkillSearchResult struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Owner         string   `json:"owner"`
	Downloads     int64    `json:"downloads"`
	LatestVersion string   `json:"latest_version"`
	UpdatedAt     string   `json:"updated_at"`
	Tags          []string `json:"tags"`
}
