package storage

import (
	"context"
	"io"
)

// Storage defines the interface for object storage operations.
type Storage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, objectName string) error
	GetURL(objectName string) string
}
