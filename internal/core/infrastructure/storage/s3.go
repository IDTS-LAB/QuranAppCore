package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/rs/zerolog"
)

// Object is a downloaded object. Body must be closed by the caller.
type Object struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

// S3 is an S3-compatible object store. It targets the dev MinIO from
// docker-compose.dev.yml by default (custom endpoint + path-style
// addressing) and works against real AWS S3 with a bucket region,
// IAM credentials, and an empty endpoint.
type S3 struct {
	client     *s3.Client
	bucket     string
	presignTTL time.Duration
	logger     zerolog.Logger
}

// New builds the client from cfg and ensures the bucket exists (creating it
// on a fresh MinIO/AWS account). Network calls are bounded by a 30s timeout.
func New(cfg config.S3Config, logger zerolog.Logger) (*S3, error) {
	presignTTL, err := cfg.PresignDuration()
	if err != nil {
		return nil, fmt.Errorf("invalid s3 presign_ttl: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
		awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				// Empty endpoint = real AWS S3: fall through to SDK default resolution.
				if cfg.Endpoint == "" {
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				}
				return aws.Endpoint{URL: cfg.Endpoint, SigningRegion: region}, nil
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Path-style (…/bucket/key) is required by MinIO; AWS S3 defaults
		// to virtual-hosted style and ignores this when false.
		o.UsePathStyle = cfg.UsePathStyle
	})

	store := &S3{client: client, bucket: cfg.Bucket, presignTTL: presignTTL, logger: logger}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

// Bucket returns the configured bucket name.
func (s *S3) Bucket() string {
	return s.bucket
}

// ensureBucket creates the bucket when a HeadBucket finds it missing, so a
// fresh MinIO works with zero manual setup. Parallel startups racing the
// create are tolerated (already-exists counts as success).
func (s *S3) ensureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket})
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return fmt.Errorf("head bucket %q: %w", s.bucket, err)
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &s.bucket})
	if err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}
	s.logger.Info().Str("bucket", s.bucket).Msg("s3 bucket ensured")
	return nil
}

// Put uploads data under key.
func (s *S3) Put(ctx context.Context, key string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	}
	if contentType != "" {
		input.ContentType = &contentType
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Get downloads the object under key. The caller must close Object.Body.
func (s *S3) Get(ctx context.Context, key string) (*Object, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	obj := &Object{Body: out.Body, Size: out.ContentLength}
	if out.ContentType != nil {
		obj.ContentType = *out.ContentType
	}
	return obj, nil
}

// Delete removes the object under key. Deleting a missing key is a no-op.
func (s *S3) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

// PresignedGetURL returns a time-limited download URL for key, so clients can
// fetch private objects directly without the app proxying bytes. TTL comes
// from s3.presign_ttl unless ttlOverride is positive.
func (s *S3) PresignedGetURL(ctx context.Context, key string, ttlOverride time.Duration) (string, error) {
	ttl := ttlOverride
	if ttl <= 0 {
		ttl = s.presignTTL
	}
	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", key, err)
	}
	return req.URL, nil
}

// PresignedGetURLs presigns a download URL for every key. Presigning is local
// crypto with no network I/O per URL, so batching is a plain loop and stays
// fast even for large batches. It fails fast on the first bad key; no URLs
// are returned in that case.
func (s *S3) PresignedGetURLs(ctx context.Context, keys []string, ttlOverride time.Duration) (map[string]string, error) {
	urls := make(map[string]string, len(keys))
	for _, key := range keys {
		url, err := s.PresignedGetURL(ctx, key, ttlOverride)
		if err != nil {
			return nil, err
		}
		urls[key] = url
	}
	return urls, nil
}

// PresignedPutURL returns a time-limited upload URL for key, so clients can
// upload directly without the app proxying bytes. When contentType is set it
// becomes part of the signature: the uploader MUST send that exact
// Content-Type header or the upload is rejected. TTL comes from
// s3.presign_ttl unless ttlOverride is positive.
func (s *S3) PresignedPutURL(ctx context.Context, key, contentType string, ttlOverride time.Duration) (string, error) {
	ttl := ttlOverride
	if ttl <= 0 {
		ttl = s.presignTTL
	}
	input := &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}
	if contentType != "" {
		input.ContentType = &contentType
	}
	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign upload %q: %w", key, err)
	}
	return req.URL, nil
}

// PresignedPutURLs presigns an upload URL for every key, all sharing one
// content-type constraint (pass "" for none). Same fail-fast semantics as
// PresignedGetURLs.
func (s *S3) PresignedPutURLs(ctx context.Context, keys []string, contentType string, ttlOverride time.Duration) (map[string]string, error) {
	urls := make(map[string]string, len(keys))
	for _, key := range keys {
		url, err := s.PresignedPutURL(ctx, key, contentType, ttlOverride)
		if err != nil {
			return nil, err
		}
		urls[key] = url
	}
	return urls, nil
}

func isNotFound(err error) bool {
	var apiErr smithy.APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	code := apiErr.ErrorCode()
	return code == "NotFound" || code == "NoSuchBucket"
}

func isAlreadyExists(err error) bool {
	var apiErr smithy.APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	code := apiErr.ErrorCode()
	return code == "BucketAlreadyExists" || code == "BucketAlreadyOwnedByYou"
}

func asAPIError(err error, target *smithy.APIError) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	*target = apiErr
	return true
}
