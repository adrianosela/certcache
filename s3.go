package certcache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"golang.org/x/crypto/acme/autocert"
)

// S3 represents an AWS S3 implementation of autocert.Cache
type S3 struct {
	client  s3iface.S3API
	bucket  string
	timeout time.Duration
}

const (
	defaultS3CertCacheBucketName   = "certcache"
	defaultS3CertCacheBucketRegion = "us-west-2"
	defaultS3CertCacheTimeout      = 10 * time.Second
)

// NewS3 returns an S3 certificate cache
func NewS3(credentials *credentials.Credentials, bucket, region string) *S3 {
	if region == "" {
		region = defaultS3CertCacheBucketRegion
	}
	if bucket == "" {
		bucket = defaultS3CertCacheBucketName
	}
	svc := s3.New(session.New(), &aws.Config{
		Credentials: credentials,
		Region:      aws.String(region),
		HTTPClient:  &http.Client{Timeout: defaultS3CertCacheTimeout},
	})
	return &S3{
		bucket: bucket,
		client: svc,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (s *S3) Get(ctx context.Context, key string) ([]byte, error) {
	results, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return nil, autocert.ErrCacheMiss
			case s3.ErrCodeNoSuchKey:
				return nil, autocert.ErrCacheMiss
			}
		}
		return nil, err
	}
	defer results.Body.Close()

	data, err := ioutil.ReadAll(results.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body for %s: %s", key, err)
	}

	return data, nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (s *S3) Put(ctx context.Context, key string, data []byte) error {
	if _, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	}); err != nil {
		return fmt.Errorf("failed to store object %s: %s", key, err)
	}
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (s *S3) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("failed to delete object %s: %s", key, err)
	}
	return nil
}
