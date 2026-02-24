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

// Patterns to exclude from bundles.
var excludePatterns = []string{
	".git",
	".DS_Store",
	"node_modules",
	"__pycache__",
	".env",
}

func shouldExclude(name string) bool {
	base := filepath.Base(name)
	for _, p := range excludePatterns {
		if base == p {
			return true
		}
		if strings.HasSuffix(p, "/") && strings.HasPrefix(name, p) {
			return true
		}
	}
	if strings.HasSuffix(base, ".pyc") {
		return true
	}
	return false
}

// Pack creates a .tar.gz bundle from the given directory and returns the temp file path.
func Pack(dir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "agentskills-bundle-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tmpFile.Close()

	gw := gzip.NewWriter(tmpFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return err
		}

		if shouldExclude(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("creating tar header: %w", err)
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("writing tar header: %w", err)
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("writing file to tar: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// Unpack extracts a .tar.gz bundle to the given destination directory.
func Unpack(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
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

		target := filepath.Join(destDir, header.Name)

		// Path traversal protection
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("path traversal detected: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("creating parent directory: %w", err)
			}
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("writing file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}
