package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"golang/config"
	"golang/pkg/logging"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client      *minio.Client
	cfg         config.MinIOConfig
	projectName string
}

// NewMinIOStorage creates a pure MinIO storage client (no local fallback).
func NewMinIOStorage(cfg config.MinIOConfig) *MinIOStorage {
	if cfg.Bucket == "" {
		cfg.Bucket = "project"
	}
	projectName := strings.Trim(cfg.ProjectName, "/")
	if projectName == "" {
		projectName = "task-management"
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		logging.Warn("failed to initialize MinIO client", logging.Err(err))
	} else {
		logging.Info("MinIO storage initialized",
			slog.String("endpoint", cfg.Endpoint),
			slog.String("bucket", cfg.Bucket),
			slog.String("project", projectName))
	}

	return &MinIOStorage{
		client:      client,
		cfg:         cfg,
		projectName: projectName,
	}
}

// Upload sends an object directly to MinIO.
func (s *MinIOStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("MinIO storage client is not connected")
	}

	cleanFileName := objectName
	if idx := strings.LastIndex(objectName, "/"); idx != -1 {
		cleanFileName = objectName[idx+1:]
	}

	// S3 key structure: {projectName}/{filename} inside bucket {cfg.Bucket}
	objectKey := fmt.Sprintf("%s/%s", s.projectName, cleanFileName)

	_, err := s.client.PutObject(ctx, s.cfg.Bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to MinIO: %w", err)
	}

	return s.GetURL(cleanFileName), nil
}

// Delete removes an object directly from MinIO.
func (s *MinIOStorage) Delete(ctx context.Context, objectName string) error {
	if s.client == nil {
		return fmt.Errorf("MinIO storage client is not connected")
	}

	cleanFileName := objectName
	if idx := strings.LastIndex(objectName, "/"); idx != -1 {
		cleanFileName = objectName[idx+1:]
	}

	objectKey := fmt.Sprintf("%s/%s", s.projectName, cleanFileName)

	if err := s.client.RemoveObject(ctx, s.cfg.Bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object from MinIO: %w", err)
	}

	return nil
}

// GetURL constructs the public URL for an object stored in MinIO.
func (s *MinIOStorage) GetURL(objectName string) string {
	cleanFileName := objectName
	if idx := strings.LastIndex(objectName, "/"); idx != -1 {
		cleanFileName = objectName[idx+1:]
	}

	publicBase := strings.TrimRight(s.cfg.PublicURL, "/")
	return fmt.Sprintf("%s/%s/%s", publicBase, s.projectName, cleanFileName)
}
