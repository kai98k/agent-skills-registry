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

// Excluded patterns when packing a skill directory.
var excludedPatterns = []string{
	".git",
	".DS_Store",
	"node_modules",
	"__pycache__",
	".env",
}

// Pack creates a .tar.gz from a skill directory.
func Pack(srcDir, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcDir, err = filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip excluded patterns
		for _, pattern := range excludedPatterns {
			base := filepath.Base(path)
			if base == pattern || strings.HasSuffix(base, ".pyc") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("file info header: %w", err)
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}

		return nil
	})
}
