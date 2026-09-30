package local

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github/OfrenDialsa/go-gin-starter/config"
)

type LocalStorageManager struct {
	basePath  string
	publicURL string
}

func NewLocalStorageManager(env *config.EnvironmentVariable) (*LocalStorageManager, error) {
	basePath := env.Storage.Local.BasePath
	if basePath == "" {
		basePath = "./storage/public"
	}

	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	return &LocalStorageManager{
		basePath:  basePath,
		publicURL: env.Storage.Local.PublicUrl,
	}, nil
}

func (l *LocalStorageManager) UploadFile(ctx context.Context, fileName string, buffer *bytes.Buffer) error {
	fullPath := filepath.Join(l.basePath, fileName)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, buffer)
	return err
}

func (l *LocalStorageManager) GetPublicURL(fileName string) (string, error) {
	if l.publicURL == "" {
		return "", errors.New("local base URL is not configured")
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(l.publicURL, "/"), strings.TrimPrefix(fileName, "/")), nil
}

func (l *LocalStorageManager) FileExists(ctx context.Context, filePath string) (bool, error) {
	fullPath := filepath.Join(l.basePath, filePath)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LocalStorageManager) GetFileSize(ctx context.Context, filePath string) (int64, error) {
	fullPath := filepath.Join(l.basePath, filePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (l *LocalStorageManager) DeleteFile(ctx context.Context, filePath string) error {
	cleanPath := filePath
	if l.publicURL != "" && strings.HasPrefix(filePath, l.publicURL) {
		key := strings.TrimPrefix(filePath, l.publicURL)
		cleanPath = strings.TrimPrefix(key, "/")
	}

	fullPath := filepath.Join(l.basePath, cleanPath)
	err := os.Remove(fullPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (l *LocalStorageManager) HealthCheck(ctx context.Context) error {
	_, err := os.Stat(l.basePath)
	return err
}
