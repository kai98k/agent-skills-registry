package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Init() error {
	return os.MkdirAll(l.basePath, 0755)
}

func (l *LocalStorage) HealthCheck(_ context.Context) error {
	_, err := os.Stat(l.basePath)
	return err
}

func (l *LocalStorage) Upload(_ context.Context, key string, reader io.Reader, _ int64) error {
	path := filepath.Join(l.basePath, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (l *LocalStorage) Download(_ context.Context, key string) (io.ReadCloser, int64, error) {
	path := filepath.Join(l.basePath, key)
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

func (l *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	path := filepath.Join(l.basePath, key)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (l *LocalStorage) Delete(_ context.Context, key string) error {
	path := filepath.Join(l.basePath, key)
	return os.Remove(path)
}
