package adapters

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dtalabs/go-packages/storage"
)

type Store struct {
	bucket, prefix string
	client         *awss3.Client
	presigner      *awss3.PresignClient
}

func New(ctx context.Context, storageURL string, regions ...string) (*Store, error) {
	if len(regions) > 1 {
		return nil, fmt.Errorf("expected at most one fallback region")
	}
	region := ""
	if len(regions) == 1 {
		region = regions[0]
	}
	settings, err := parseStorageConfig(storageURL, region)
	if err != nil {
		return nil, err
	}
	options := []func(*awsconfig.LoadOptions) error{}
	if settings.Region != "" {
		options = append(options, awsconfig.WithRegion(settings.Region))
	}
	if settings.AccessKey != "" {
		credentialProvider := credentials.NewStaticCredentialsProvider(settings.AccessKey, settings.SecretKey, "")
		options = append(options, awsconfig.WithCredentialsProvider(credentialProvider))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}
	client := awss3.NewFromConfig(cfg, func(options *awss3.Options) {
		if settings.Endpoint != "" {
			options.BaseEndpoint = aws.String(settings.Endpoint)
			options.UsePathStyle = true
		}
	})
	return &Store{bucket: settings.Bucket, prefix: settings.Prefix, client: client, presigner: awss3.NewPresignClient(client)}, nil
}

func (s *Store) Configured() bool { return s != nil && s.bucket != "" }

func (s *Store) ObjectKey(userID, uploadID, filename string) string {
	ext := path.Ext(filename)
	if len(ext) > 12 {
		ext = ""
	}
	key := path.Join("uploads", userID, uploadID+strings.ToLower(ext))
	if s.prefix != "" {
		key = path.Join(s.prefix, key)
	}
	return key
}

func (s *Store) PresignPut(ctx context.Context, key, contentType string) (string, error) {
	if !s.Configured() {
		return "", storage.ErrNotConfigured
	}
	result, err := s.presigner.PresignPutObject(ctx, &awss3.PutObjectInput{Bucket: &s.bucket, Key: &key, ContentType: &contentType}, awss3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}
	return result.URL, nil
}

func (s *Store) PresignGet(ctx context.Context, key string) (string, error) {
	if !s.Configured() {
		return "", storage.ErrNotConfigured
	}
	result, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{Bucket: &s.bucket, Key: &key}, awss3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (s *Store) Exists(ctx context.Context, key string) error {
	if !s.Configured() {
		return storage.ErrNotConfigured
	}
	_, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{Bucket: &s.bucket, Key: &key})
	return err
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if !s.Configured() {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	return err
}

var _ storage.Store = (*Store)(nil)
