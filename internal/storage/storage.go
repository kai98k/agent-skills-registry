package storage

import (
	"context"
	"io"
)

// Storage defines the interface for bundle file storage.
// Implemented by LocalStorage and S3Storage.
type Storage interface {
	Init() error
	HealthCheck(ctx context.Context) error
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
}
