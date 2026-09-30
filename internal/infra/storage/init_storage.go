package storage

import (
	"context"
	"errors"
	"github/OfrenDialsa/go-gin-starter/config"

	awss3 "github/OfrenDialsa/go-gin-starter/internal/infra/storage/aws_s3"
	local "github/OfrenDialsa/go-gin-starter/internal/infra/storage/local"
	"github/OfrenDialsa/go-gin-starter/internal/infra/storage/minio"
)

func NewStorageService(ctx context.Context, env *config.EnvironmentVariable) (StorageService, error) {
	switch env.Storage.Type {
	case "minio":
		return minio.NewMinioStorageManager(env)
	case "s3":
		return awss3.NewS3StorageManager(ctx, env)
	case "local":
		return local.NewLocalStorageManager(env)
	default:
		return nil, errors.New("unsupported storage driver")
	}
}
