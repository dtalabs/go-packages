package adapters

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dtalabs/go-packages/storage"
)

type Store struct {
	bucket, prefix string
	client         *awss3.Client
	presigner      *awss3.PresignClient
}

func New(ctx context.Context, storageURL, region string) (*Store, error) {
	bucket, prefix, endpoint := "", "", ""
	if storageURL != "" {
		parsed, err := url.Parse(storageURL)
		if err != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid STORAGE_URL")
		}
		switch parsed.Scheme {
		case "s3":
			bucket, prefix = parsed.Host, strings.Trim(parsed.Path, "/")
		case "http", "https":
			parts := strings.SplitN(strings.Trim(parsed.Path, "/"), "/", 2)
			if len(parts) == 0 || parts[0] == "" {
				return nil, fmt.Errorf("STORAGE_URL must include a bucket after the endpoint")
			}
			bucket = parts[0]
			if len(parts) == 2 {
				prefix = strings.Trim(parts[1], "/")
			}
			endpoint = parsed.Scheme + "://" + parsed.Host
		default:
			return nil, fmt.Errorf("STORAGE_URL must use s3://bucket/prefix, http://endpoint/bucket/prefix, or https://endpoint/bucket/prefix")
		}
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := awss3.NewFromConfig(cfg, func(options *awss3.Options) {
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
			options.UsePathStyle = true
		}
	})
	return &Store{bucket: bucket, prefix: prefix, client: client, presigner: awss3.NewPresignClient(client)}, nil
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
