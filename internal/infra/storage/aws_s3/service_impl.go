package awss3

import (
	"bytes"
	"context"
	"errors"
	"github/OfrenDialsa/go-gin-starter/config"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
)

type S3StorageManager struct {
	PublicUrl string
	Bucket    string

	Client *s3.Client
}

func NewS3StorageManager(ctx context.Context, env *config.EnvironmentVariable) (*S3StorageManager, error) {
	scheme := "https"
	if !env.Storage.S3.SSL {
		scheme = "http"
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               scheme + "://" + env.Storage.S3.Endpoint,
				SigningRegion:     env.Storage.S3.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := aws_cfg.LoadDefaultConfig(
		ctx,
		aws_cfg.WithRegion(env.Storage.S3.Region),
		aws_cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(env.Storage.S3.AccessKey, env.Storage.S3.SecretKey, "")),
		aws_cfg.WithEndpointResolverWithOptions(resolver),
		aws_cfg.WithClientLogMode(aws.LogRetries),
		aws_cfg.WithLogger(logging.NewStandardLogger(nil)),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = env.Storage.S3.UsePathStyle
	})

	return &S3StorageManager{
		PublicUrl: env.Storage.S3.PublicUrl,
		Bucket:    env.Storage.S3.Bucket,
		Client:    client,
	}, nil
}

func (s *S3StorageManager) UploadFile(ctx context.Context, fileName string, buffer *bytes.Buffer) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	return err
}

func (s *S3StorageManager) GetPublicURL(fileName string) (string, error) {
	return s.PublicUrl + "/" + s.Bucket + "/" + fileName, nil
}

func (s *S3StorageManager) FileExists(ctx context.Context, filePath string) (bool, error) {
	_, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3StorageManager) GetFileSize(ctx context.Context, filePath string) (int64, error) {
	output, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return 0, err
	}
	return aws.ToInt64(output.ContentLength), nil
}

func (s *S3StorageManager) DeleteFile(ctx context.Context, filePath string) error {
	prefix := s.PublicUrl + "/" + s.Bucket + "/"
	cleanPath := strings.TrimPrefix(filePath, prefix)

	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(cleanPath),
	})
	return err
}

func (s *S3StorageManager) HealthCheck(ctx context.Context) error {
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.Bucket),
	})
	return err
}
