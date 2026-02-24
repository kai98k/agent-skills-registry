package bundle

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// MaxUnpackSize is the maximum total size after decompression (200MB).
	MaxUnpackSize int64 = 200 * 1024 * 1024
)

// Unpack extracts a .tar.gz to a destination directory.
// It enforces path traversal protection and zip bomb limits.
func Unpack(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	return UnpackReader(f, destDir)
}

// UnpackReader extracts a .tar.gz from a reader to a destination directory.
func UnpackReader(r io.Reader, destDir string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var totalSize int64

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		// Path traversal protection
		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), destDir) {
			return fmt.Errorf("path traversal detected: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}

		case tar.TypeReg:
			// Zip bomb protection
			totalSize += header.Size
			if totalSize > MaxUnpackSize {
				return fmt.Errorf("archive exceeds maximum unpack size (%d MB)", MaxUnpackSize/(1024*1024))
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent: %w", err)
			}

			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}

			// Also limit individual file copy with a capped reader
			if _, err := io.Copy(outFile, io.LimitReader(tr, header.Size+1)); err != nil {
				outFile.Close()
				return fmt.Errorf("write file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}

// FindSKILLMD searches for SKILL.md in the unpacked directory (root or one level deep).
func FindSKILLMD(dir string) (string, error) {
	// Check root
	rootPath := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(rootPath); err == nil {
		return rootPath, nil
	}

	// Check one level deep
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			subPath := filepath.Join(dir, e.Name(), "SKILL.md")
			if _, err := os.Stat(subPath); err == nil {
				return subPath, nil
			}
		}
	}

	return "", fmt.Errorf("SKILL.md not found in archive")
}
