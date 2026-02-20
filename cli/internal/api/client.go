package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/liuyukai/agentskills-cli/internal/config"
)

// Client wraps HTTP calls to the AgentSkills API
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// NewClient creates a new API client from config
func NewClient() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL: strings.TrimRight(cfg.APIURL, "/"),
		Token:   cfg.Token,
		HTTP:    &http.Client{},
	}, nil
}

// PublishResponse represents the API publish response
type PublishResponse struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Checksum    string   `json:"checksum"`
	PublishedAt string   `json:"published_at"`
	Providers   []string `json:"providers"`
}

// SkillInfo represents skill info response
type SkillInfo struct {
	Name          string         `json:"name"`
	Owner         string         `json:"owner"`
	Downloads     int64          `json:"downloads"`
	CreatedAt     string         `json:"created_at"`
	LatestVersion *VersionDetail `json:"latest_version"`
}

// VersionDetail represents a version's details
type VersionDetail struct {
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Checksum    string            `json:"checksum"`
	SizeBytes   int64             `json:"size_bytes"`
	PublishedAt string            `json:"published_at"`
	Providers   []string          `json:"providers"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// VersionSummary represents a version summary
type VersionSummary struct {
	Version     string   `json:"version"`
	Checksum    string   `json:"checksum"`
	SizeBytes   int64    `json:"size_bytes"`
	PublishedAt string   `json:"published_at"`
	Providers   []string `json:"providers"`
}

// VersionsResponse represents the list versions response
type VersionsResponse struct {
	Name     string           `json:"name"`
	Versions []VersionSummary `json:"versions"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Owner         string   `json:"owner"`
	Downloads     int64    `json:"downloads"`
	LatestVersion string   `json:"latest_version"`
	UpdatedAt     string   `json:"updated_at"`
	Tags          []string `json:"tags"`
	Providers     []string `json:"providers"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
	Results []SearchResult `json:"results"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// Publish uploads a skill bundle
func (c *Client) Publish(bundleData []byte, filename string, providers string) (*PublishResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := part.Write(bundleData); err != nil {
		return nil, fmt.Errorf("writing bundle data: %w", err)
	}

	if providers != "" {
		if err := writer.WriteField("providers", providers); err != nil {
			return nil, fmt.Errorf("writing providers field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/skills/publish", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("publishing skill: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		var apiErr ErrorResponse
		json.Unmarshal(respBody, &apiErr)
		msg := apiErr.Detail
		if msg == "" {
			msg = apiErr.Error
		}
		return nil, fmt.Errorf("publish failed (%d): %s", resp.StatusCode, msg)
	}

	var result PublishResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &result, nil
}

// GetSkill retrieves skill info
func (c *Client) GetSkill(name string) (*SkillInfo, error) {
	resp, err := c.HTTP.Get(c.BaseURL + "/v1/skills/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var info SkillInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Download downloads a skill bundle and returns the raw bytes plus checksum header
func (c *Client) Download(name, version string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/v1/skills/%s/versions/%s/download", c.BaseURL, name, version)
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, "", fmt.Errorf("skill '%s@%s' not found", name, version)
	}
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	checksum := resp.Header.Get("X-Checksum-SHA256")
	return data, checksum, nil
}

// Search searches for skills
func (c *Client) Search(query, tag, provider string, page, perPage int) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/v1/skills?page=%d&per_page=%d", c.BaseURL, page, perPage)
	if query != "" {
		url += "&q=" + query
	}
	if tag != "" {
		url += "&tag=" + tag
	}
	if provider != "" {
		url += "&provider=" + provider
	}

	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search failed: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
