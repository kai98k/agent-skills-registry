package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExcludePatterns lists files/dirs to exclude from bundles
var ExcludePatterns = []string{
	".git",
	".DS_Store",
	"node_modules",
	"__pycache__",
	".env",
}

// shouldExclude checks if a path should be excluded
func shouldExclude(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range ExcludePatterns {
		if base == pattern {
			return true
		}
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(base, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}
	// Exclude .pyc files
	if strings.HasSuffix(base, ".pyc") {
		return true
	}
	return false
}

// Pack creates a .tar.gz from a directory and returns the bytes
func Pack(dir string) ([]byte, error) {
	// Create a pipe to stream tar.gz
	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		gw := gzip.NewWriter(pw)
		tw := tar.NewWriter(gw)

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			if relPath == "." {
				return nil
			}

			if shouldExclude(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !info.IsDir() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				if _, err := io.Copy(tw, f); err != nil {
					return err
				}
			}

			return nil
		})

		tw.Close()
		gw.Close()
		pw.CloseWithError(err)
		errCh <- err
	}()

	data, readErr := io.ReadAll(pr)
	walkErr := <-errCh

	if walkErr != nil {
		return nil, fmt.Errorf("packing bundle: %w", walkErr)
	}
	if readErr != nil {
		return nil, fmt.Errorf("reading packed data: %w", readErr)
	}

	return data, nil
}

// Unpack extracts a .tar.gz to a target directory
func Unpack(data []byte, targetDir string) error {
	r := strings.NewReader(string(data))
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		// Validate path (prevent traversal)
		target := filepath.Join(targetDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(targetDir)) {
			return fmt.Errorf("path traversal detected: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

// Checksum computes SHA-256 hex digest of data
func Checksum(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
