package storage

import (
	"bytes"
	"context"
)

type StorageService interface {
	UploadFile(ctx context.Context, fileName string, buffer *bytes.Buffer) error
	GetPublicURL(fileName string) (string, error)

	FileExists(ctx context.Context, filePath string) (bool, error)
	GetFileSize(ctx context.Context, filePath string) (int64, error)
	DeleteFile(ctx context.Context, filePath string) error

	HealthCheck(ctx context.Context) error
}
