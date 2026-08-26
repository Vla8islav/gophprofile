package domain

import (
	"context"
	"io"
)

// FileStorage abstracts S3
type FileStorage interface {
	Upload(ctx context.Context, key string, contentType string, size int64, content io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
