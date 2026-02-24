package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *Client) doRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	u := c.BaseURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.HTTPClient.Do(req)
}

// Publish uploads a .tar.gz bundle to the registry.
func (c *Client) Publish(archivePath string) (map[string]interface{}, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(archivePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	w.Close()

	resp, err := c.doRequest("POST", "/v1/skills/publish", &buf, w.FormDataContentType())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusCreated {
		errMsg := "unknown error"
		if e, ok := result["error"].(string); ok {
			errMsg = e
		}
		return nil, fmt.Errorf("publish failed (%d): %s", resp.StatusCode, errMsg)
	}
	return result, nil
}

// GetSkill fetches skill info.
func (c *Client) GetSkill(name string) (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/v1/skills/"+url.PathEscape(name), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusOK {
		errMsg := "unknown error"
		if e, ok := result["error"].(string); ok {
			errMsg = e
		}
		return nil, fmt.Errorf("get skill failed (%d): %s", resp.StatusCode, errMsg)
	}
	return result, nil
}

// Download downloads a specific version bundle and returns the reader, checksum header, and content length.
func (c *Client) Download(name, version string) (io.ReadCloser, string, int64, error) {
	path := fmt.Sprintf("/v1/skills/%s/versions/%s/download", url.PathEscape(name), url.PathEscape(version))
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, "", 0, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", 0, fmt.Errorf("download failed (%d)", resp.StatusCode)
	}

	checksum := resp.Header.Get("X-Checksum-SHA256")
	return resp.Body, checksum, resp.ContentLength, nil
}

// Search searches for skills.
func (c *Client) Search(query string) (map[string]interface{}, error) {
	path := "/v1/skills?q=" + url.QueryEscape(query)
	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
