package storage

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"emc_lb/src/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AvatarStorage interface {
	EnsureBucket(context.Context) error
	UploadAvatar(context.Context, string, string, []byte, string) (string, error)
}

type LocalStackS3Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

func NewLocalStackS3Storage(ctx context.Context) (*LocalStackS3Storage, error) {
	region := utils.GetEnv("AWS_REGION", "us-east-1")
	endpoint := utils.GetEnv("LOCALSTACK_ENDPOINT", "http://localhost:4566")
	accessKeyID := utils.GetEnv("AWS_ACCESS_KEY_ID", "test")
	secretAccessKey := utils.GetEnv("AWS_SECRET_ACCESS_KEY", "test")
	bucket := utils.GetEnv("S3_AVATAR_BUCKET", "emc-lb-avatars")

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
		options.BaseEndpoint = aws.String(endpoint)
	})

	return &LocalStackS3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: strings.TrimRight(endpoint, "/"),
	}, nil
}

func (s *LocalStackS3Storage) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	return err
}

func (s *LocalStackS3Storage) UploadAvatar(ctx context.Context, userID string, fileName string, fileData []byte, contentType string) (string, error) {
	objectKey := buildAvatarObjectKey(userID, fileName)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectKey), nil
}

func buildAvatarObjectKey(userID string, fileName string) string {
	safeUserID := strings.NewReplacer("/", "-", "\\", "-").Replace(strings.ToLower(userID))
	extension := filepath.Ext(fileName)
	if extension == "" {
		extension = ".bin"
	}

	return fmt.Sprintf("avatars/%s%s", safeUserID, extension)
}
