package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liuyukai/agentskills/migrations"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type SQLiteDB struct {
	db   *sql.DB
	path string
}

func NewSQLite(dsn string) *SQLiteDB {
	return &SQLiteDB{path: dsn}
}

func (s *SQLiteDB) Open() error {
	db, err := sql.Open("sqlite3", s.path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	s.db = db
	return nil
}

func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteDB) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLiteDB) Migrate() error {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, name := range sqlFiles {
		data, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := s.db.Exec(string(data)); err != nil {
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
	}
	return nil
}

// ── Users ──────────────────────────────────────────

func (s *SQLiteDB) GetUserByToken(ctx context.Context, token string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, api_token, created_at FROM users WHERE api_token = ?", token,
	).Scan(&u.ID, &u.Username, &u.APIToken, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *SQLiteDB) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, api_token, created_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Username, &u.APIToken, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *SQLiteDB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, api_token, created_at FROM users WHERE username = ?", username,
	).Scan(&u.ID, &u.Username, &u.APIToken, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *SQLiteDB) CreateUser(ctx context.Context, username, apiToken string) (*User, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO users (id, username, api_token, created_at) VALUES (?, ?, ?, ?)",
		id, username, apiToken, now,
	)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, Username: username, APIToken: apiToken, CreatedAt: now}, nil
}

// ── Tokens (PAT) ──────────────────────────────────

func (s *SQLiteDB) CreateToken(ctx context.Context, userID, name, tokenHash, tokenPrefix string, expiresAt *time.Time) (*Token, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO tokens (id, user_id, name, token_hash, token_prefix, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, userID, name, tokenHash, tokenPrefix, expiresAt, now,
	)
	if err != nil {
		return nil, err
	}
	return &Token{
		ID:          id,
		UserID:      userID,
		Name:        name,
		TokenHash:   tokenHash,
		TokenPrefix: tokenPrefix,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
	}, nil
}

func (s *SQLiteDB) GetUserByTokenHash(ctx context.Context, tokenHash string) (*User, error) {
	var t Token
	err := s.db.QueryRowContext(ctx,
		"SELECT id, user_id, expires_at FROM tokens WHERE token_hash = ?", tokenHash,
	).Scan(&t.ID, &t.UserID, &t.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Check expiration
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil
	}
	return s.GetUserByID(ctx, t.UserID)
}

func (s *SQLiteDB) ListTokens(ctx context.Context, userID string) ([]Token, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, user_id, name, token_prefix, expires_at, last_used_at, created_at FROM tokens WHERE user_id = ? ORDER BY created_at DESC", userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.ExpiresAt, &t.LastUsedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (s *SQLiteDB) DeleteToken(ctx context.Context, tokenID, userID string) error {
	res, err := s.db.ExecContext(ctx,
		"DELETE FROM tokens WHERE id = ? AND user_id = ?", tokenID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLiteDB) UpdateTokenLastUsed(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE tokens SET last_used_at = ? WHERE token_hash = ?", time.Now().UTC(), tokenHash,
	)
	return err
}

// ── Skills ─────────────────────────────────────────

func (s *SQLiteDB) GetSkillByName(ctx context.Context, name string) (*Skill, error) {
	sk := &Skill{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, owner_id, downloads, created_at, updated_at FROM skills WHERE name = ?", name,
	).Scan(&sk.ID, &sk.Name, &sk.OwnerID, &sk.Downloads, &sk.CreatedAt, &sk.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sk, err
}

func (s *SQLiteDB) CreateSkill(ctx context.Context, name, ownerID string) (*Skill, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO skills (id, name, owner_id, downloads, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)",
		id, name, ownerID, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Skill{ID: id, Name: name, OwnerID: ownerID, Downloads: 0, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *SQLiteDB) IncrementDownloads(ctx context.Context, skillID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE skills SET downloads = downloads + 1 WHERE id = ?", skillID,
	)
	return err
}

func (s *SQLiteDB) UpdateSkillTimestamp(ctx context.Context, skillID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE skills SET updated_at = ? WHERE id = ?", time.Now().UTC(), skillID,
	)
	return err
}

func (s *SQLiteDB) SearchSkills(ctx context.Context, query, tag string, page, perPage int) ([]SkillSearchResult, int, error) {
	// Simple approach: query skills + latest version in two steps
	var conditions []string
	var countArgs []interface{}

	if query != "" {
		conditions = append(conditions, "s.name LIKE ?")
		countArgs = append(countArgs, "%"+query+"%")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM skills s %s", where)
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch skills
	offset := (page - 1) * perPage
	searchArgs := append([]interface{}{}, countArgs...)
	searchArgs = append(searchArgs, perPage, offset)

	searchQuery := fmt.Sprintf(`
		SELECT s.id, s.name, s.downloads, s.updated_at, u.username
		FROM skills s
		JOIN users u ON u.id = s.owner_id
		%s
		ORDER BY s.downloads DESC, s.updated_at DESC
		LIMIT ? OFFSET ?`, where)

	rows, err := s.db.QueryContext(ctx, searchQuery, searchArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []SkillSearchResult
	for rows.Next() {
		var r SkillSearchResult
		var skillID string
		var updatedAt time.Time
		if err := rows.Scan(&skillID, &r.Name, &r.Downloads, &updatedAt, &r.Owner); err != nil {
			return nil, 0, err
		}
		r.UpdatedAt = updatedAt.Format(time.RFC3339)

		// Get latest version metadata
		latest, _ := s.GetLatestVersion(ctx, skillID)
		if latest != nil {
			r.LatestVersion = latest.Version
			if desc, ok := latest.Metadata["description"].(string); ok {
				r.Description = desc
			}
			if tags, ok := latest.Metadata["tags"].([]interface{}); ok {
				for _, t := range tags {
					if ts, ok := t.(string); ok {
						r.Tags = append(r.Tags, ts)
					}
				}
			}

			// Filter by tag if specified
			if tag != "" {
				found := false
				for _, t := range r.Tags {
					if t == tag {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Filter by description if query specified
			if query != "" && !strings.Contains(r.Name, query) {
				if !strings.Contains(strings.ToLower(r.Description), strings.ToLower(query)) {
					continue
				}
			}
		}

		results = append(results, r)
	}
	return results, total, rows.Err()
}

// ── Versions ───────────────────────────────────────

func (s *SQLiteDB) CreateVersion(ctx context.Context, v *SkillVersion) error {
	metaJSON, err := json.Marshal(v.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	v.PublishedAt = time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO skill_versions (id, skill_id, version, bundle_key, metadata, checksum, size_bytes, published_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		v.ID, v.SkillID, v.Version, v.BundleKey, string(metaJSON), v.Checksum, v.SizeBytes, v.PublishedAt,
	)
	return err
}

func (s *SQLiteDB) GetVersion(ctx context.Context, skillID, version string) (*SkillVersion, error) {
	v := &SkillVersion{}
	var metaJSON string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, skill_id, version, bundle_key, metadata, checksum, size_bytes, published_at FROM skill_versions WHERE skill_id = ? AND version = ?",
		skillID, version,
	).Scan(&v.ID, &v.SkillID, &v.Version, &v.BundleKey, &metaJSON, &v.Checksum, &v.SizeBytes, &v.PublishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(metaJSON), &v.Metadata)
	return v, nil
}

func (s *SQLiteDB) GetLatestVersion(ctx context.Context, skillID string) (*SkillVersion, error) {
	v := &SkillVersion{}
	var metaJSON string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, skill_id, version, bundle_key, metadata, checksum, size_bytes, published_at FROM skill_versions WHERE skill_id = ? ORDER BY published_at DESC LIMIT 1",
		skillID,
	).Scan(&v.ID, &v.SkillID, &v.Version, &v.BundleKey, &metaJSON, &v.Checksum, &v.SizeBytes, &v.PublishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(metaJSON), &v.Metadata)
	return v, nil
}

func (s *SQLiteDB) ListVersions(ctx context.Context, skillID string) ([]SkillVersion, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, skill_id, version, bundle_key, metadata, checksum, size_bytes, published_at FROM skill_versions WHERE skill_id = ? ORDER BY published_at DESC",
		skillID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []SkillVersion
	for rows.Next() {
		var v SkillVersion
		var metaJSON string
		if err := rows.Scan(&v.ID, &v.SkillID, &v.Version, &v.BundleKey, &metaJSON, &v.Checksum, &v.SizeBytes, &v.PublishedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metaJSON), &v.Metadata)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
