package minio

import (
	"bytes"
	"context"
	"github/OfrenDialsa/go-gin-starter/config"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorageManager struct {
	PublicUrl string
	Bucket    string

	Client *minio.Client
}

func NewMinioStorageManager(env *config.EnvironmentVariable) (*MinioStorageManager, error) {
	minioClient, err := minio.New(env.Storage.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(env.Storage.S3.AccessKey, env.Storage.S3.SecretKey, ""),
		Secure: env.Storage.S3.SSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioStorageManager{
		PublicUrl: env.Storage.S3.PublicUrl,
		Bucket:    env.Storage.S3.Bucket,
		Client:    minioClient,
	}, nil
}

func (m *MinioStorageManager) UploadFile(ctx context.Context, fileName string, buffer *bytes.Buffer) error {
	_, err := m.Client.PutObject(ctx, m.Bucket, fileName, buffer, int64(buffer.Len()), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *MinioStorageManager) GetPublicURL(fileName string) (string, error) {
	return r.PublicUrl + "/" + r.Bucket + "/" + fileName, nil
}

func (m *MinioStorageManager) FileExists(ctx context.Context, filePath string) (bool, error) {
	_, err := m.Client.StatObject(ctx, m.Bucket, filePath, minio.StatObjectOptions{})
	if err != nil {
		if errResp := minio.ToErrorResponse(err); errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioStorageManager) GetFileSize(ctx context.Context, filePath string) (int64, error) {
	objInfo, err := m.Client.StatObject(ctx, m.Bucket, filePath, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return objInfo.Size, nil
}

func (m *MinioStorageManager) DeleteFile(ctx context.Context, filePath string) error {
	prefix := m.PublicUrl + "/" + m.Bucket + "/"
	cleanPath := strings.TrimPrefix(filePath, prefix)

	return m.Client.RemoveObject(ctx, m.Bucket, cleanPath, minio.RemoveObjectOptions{})
}

func (m *MinioStorageManager) HealthCheck(ctx context.Context) error {
	_, err := m.Client.BucketExists(ctx, m.Bucket)
	return err
}
