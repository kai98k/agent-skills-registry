package database

import (
	"context"
	"time"
)

// Database defines the interface for all database operations.
// Implemented by SQLiteDB and PostgresDB.
type Database interface {
	// Lifecycle
	Open() error
	Close() error
	Migrate() error
	HealthCheck(ctx context.Context) error

	// Users
	GetUserByToken(ctx context.Context, token string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, username, apiToken string) (*User, error)

	// Tokens (PAT)
	CreateToken(ctx context.Context, userID, name, tokenHash, tokenPrefix string, expiresAt *time.Time) (*Token, error)
	GetUserByTokenHash(ctx context.Context, tokenHash string) (*User, error)
	ListTokens(ctx context.Context, userID string) ([]Token, error)
	DeleteToken(ctx context.Context, tokenID, userID string) error
	UpdateTokenLastUsed(ctx context.Context, tokenHash string) error

	// Skills
	GetSkillByName(ctx context.Context, name string) (*Skill, error)
	CreateSkill(ctx context.Context, name, ownerID string) (*Skill, error)
	IncrementDownloads(ctx context.Context, skillID string) error
	UpdateSkillTimestamp(ctx context.Context, skillID string) error
	SearchSkills(ctx context.Context, query, tag string, page, perPage int) ([]SkillSearchResult, int, error)

	// Versions
	CreateVersion(ctx context.Context, v *SkillVersion) error
	GetVersion(ctx context.Context, skillID, version string) (*SkillVersion, error)
	GetLatestVersion(ctx context.Context, skillID string) (*SkillVersion, error)
	ListVersions(ctx context.Context, skillID string) ([]SkillVersion, error)
}
