// Package filestorage implements domain.FileStorage on top of any
// S3-compatible object store (MinIO, AWS S3) via the MinIO client.
package filestorage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

// NewMinioStorage builds the client and makes a best-effort attempt to create
// the bucket. When store is unreachable, /health reports s3 as down, and uploads fail per-request.
func NewMinioStorage(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client for %s: %w", endpoint, err)
	}

	storage := &MinioStorage{client: client, bucket: bucket}
	_ = storage.EnsureBucket(ctx) // best effort; see doc comment
	return storage, nil
}

func (s *MinioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket %s: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", s.bucket, err)
	}
	return nil
}

func (s *MinioStorage) Upload(ctx context.Context, key string, contentType string, size int64, content io.Reader) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, content, size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload %s to bucket %s: %w", key, s.bucket, err)
	}
	return nil
}

func (s *MinioStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download %s from bucket %s: %w", key, s.bucket, err)
	}
	// GetObject is lazy; Stat forces the request so missing keys fail here,
	// not on the first Read halfway through writing a response.
	if _, err := object.Stat(); err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("failed to stat %s in bucket %s: %w", key, s.bucket, err)
	}
	return object, nil
}

func (s *MinioStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete %s from bucket %s: %w", key, s.bucket, err)
	}
	return nil
}

func (s *MinioStorage) Ping(ctx context.Context) error {
	if _, err := s.client.BucketExists(ctx, s.bucket); err != nil {
		return fmt.Errorf("file storage unreachable: %w", err)
	}
	return nil
}
