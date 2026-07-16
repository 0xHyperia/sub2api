package repository

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3TicketStorage struct {
	active bool
	client *s3.Client
	bucket string
	expiry time.Duration
}

func ProvideTicketStorage(cfg *config.Config) (service.TicketStorage, error) {
	if !cfg.TicketStorage.Active() {
		return &s3TicketStorage{}, nil
	}
	client, err := newS3Client(context.Background(), s3ClientParams{Endpoint: cfg.TicketStorage.Endpoint, Region: cfg.TicketStorage.Region, AccessKeyID: cfg.TicketStorage.AccessKeyID, SecretAccessKey: cfg.TicketStorage.SecretAccessKey, ForcePathStyle: cfg.TicketStorage.ForcePathStyle})
	if err != nil {
		return nil, err
	}
	expiry := time.Duration(cfg.TicketStorage.PresignExpiryMinutes) * time.Minute
	if expiry <= 0 {
		expiry = 5 * time.Minute
	}
	return &s3TicketStorage{active: true, client: client, bucket: cfg.TicketStorage.Bucket, expiry: expiry}, nil
}
func (s *s3TicketStorage) Active() bool { return s != nil && s.active }
func (s *s3TicketStorage) Put(ctx context.Context, key, contentType string, body io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &key, Body: body, ContentType: &contentType, ContentLength: aws.Int64(size)})
	return err
}
func (s *s3TicketStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	return err
}
func (s *s3TicketStorage) PresignGet(ctx context.Context, key, filename, contentType string) (string, error) {
	disposition := "attachment; filename*=UTF-8''" + url.PathEscape(filename)
	out, err := s3.NewPresignClient(s.client).PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: &key, ResponseContentDisposition: &disposition, ResponseContentType: &contentType}, s3.WithPresignExpires(s.expiry))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
