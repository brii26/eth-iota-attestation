package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config holds connection options for RustFS / S3 storage.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
}

// S3Store provides methods to put and get SBOM and scan files in RustFS.
type S3Store struct {
	client *minio.Client
}

const (
	BucketSBOMs = "sboms"
	BucketScans = "scans"
)

// NewS3Store initializes minio client and ensures the required buckets exist.
func NewS3Store(ctx context.Context, cfg S3Config) (*S3Store, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://127.0.0.1:9000"
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = "rustfsadmin"
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = "rustfsdevpassword123"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 endpoint URL: %w", err)
	}

	client, err := minio.New(u.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       u.Scheme == "https",
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	store := &S3Store{client: client}

	for _, bucket := range []string{BucketSBOMs, BucketScans} {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil || !exists {
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
				// ponytail: log/return error if creation fails; RustFS might have existing bucket
				return nil, fmt.Errorf("ensure bucket %s exists: %w", bucket, err)
			}
		}
	}

	return store, nil
}

// PutSBOM uploads raw SBOM bytes to the sboms bucket.
func (s *S3Store) PutSBOM(ctx context.Context, key string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, BucketSBOMs, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("put SBOM object %s: %w", key, err)
	}
	return nil
}

// PutScan uploads raw scan bytes to the scans bucket.
func (s *S3Store) PutScan(ctx context.Context, key string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, BucketScans, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("put scan object %s: %w", key, err)
	}
	return nil
}

// GetSBOM retrieves raw SBOM bytes from the sboms bucket.
func (s *S3Store) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, BucketSBOMs, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get SBOM object %s: %w", key, err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// GetScan retrieves raw scan bytes from the scans bucket.
func (s *S3Store) GetScan(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, BucketScans, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get scan object %s: %w", key, err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}
