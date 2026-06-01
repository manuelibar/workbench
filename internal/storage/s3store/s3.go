package s3store

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/manuelibar/workbench/internal/storage"
)

type Store struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
	now     func() time.Time
}

func New(bucket string, client *s3.Client) (*Store, error) {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return nil, storage.ErrInvalid
	}
	if client == nil {
		return nil, storage.ErrInvalid
	}
	return &Store{
		bucket:  bucket,
		client:  client,
		presign: s3.NewPresignClient(client),
		now:     time.Now,
	}, nil
}

func (s *Store) PresignPut(ctx context.Context, key, contentType string, metadata map[string]string, ttl time.Duration) (storage.SignedURL, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}
	result, err := s.presign.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return storage.SignedURL{}, err
	}
	headers := headersFromHTTP(result.SignedHeader)
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	return storage.SignedURL{
		URL:       result.URL,
		Method:    http.MethodPut,
		Headers:   headers,
		ExpiresAt: s.now().Add(ttl).UTC().Format(time.RFC3339),
	}, nil
}

func (s *Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (storage.SignedURL, error) {
	result, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return storage.SignedURL{}, err
	}
	return storage.SignedURL{
		URL:       result.URL,
		Method:    http.MethodGet,
		Headers:   headersFromHTTP(result.SignedHeader),
		ExpiresAt: s.now().Add(ttl).UTC().Format(time.RFC3339),
	}, nil
}

func (s *Store) StartMultipart(ctx context.Context, key, contentType string, metadata map[string]string, partCount int, ttl time.Duration) (storage.MultipartUpload, error) {
	created, err := s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	})
	if err != nil {
		return storage.MultipartUpload{}, err
	}
	out := storage.MultipartUpload{
		Key:      key,
		UploadID: aws.ToString(created.UploadId),
		Parts:    make([]storage.MultipartPart, 0, partCount),
	}
	for i := 1; i <= partCount; i++ {
		partNumber := int32(i)
		result, err := s.presign.PresignUploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(s.bucket),
			Key:        aws.String(key),
			UploadId:   created.UploadId,
			PartNumber: aws.Int32(partNumber),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = ttl
		})
		if err != nil {
			return storage.MultipartUpload{}, err
		}
		out.Parts = append(out.Parts, storage.MultipartPart{
			PartNumber: i,
			URL:        result.URL,
			Method:     http.MethodPut,
			Headers:    headersFromHTTP(result.SignedHeader),
		})
	}
	return out, nil
}

func (s *Store) GetObject(ctx context.Context, key, byteRange string) (storage.Object, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if strings.TrimSpace(byteRange) != "" {
		input.Range = aws.String(byteRange)
	}
	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		if isNotFound(err) {
			return storage.Object{}, storage.ErrNotFound
		}
		return storage.Object{}, err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return storage.Object{}, err
	}
	return storage.Object{
		Key:         key,
		Body:        body,
		ContentType: aws.ToString(result.ContentType),
		Metadata:    result.Metadata,
	}, nil
}

func (s *Store) PutObject(ctx context.Context, key string, body []byte, contentType string, metadata map[string]string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	})
	return err
}

func (s *Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *Store) HeadObject(ctx context.Context, key string) (storage.ObjectInfo, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return storage.ObjectInfo{}, storage.ErrNotFound
		}
		return storage.ObjectInfo{}, err
	}
	return storage.ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		ETag:         strings.Trim(aws.ToString(result.ETag), "\""),
		ContentType:  aws.ToString(result.ContentType),
		Metadata:     result.Metadata,
		LastModified: aws.ToTime(result.LastModified),
	}, nil
}

func (s *Store) ListObjects(ctx context.Context, prefix string) ([]storage.ObjectInfo, error) {
	var out []storage.ObjectInfo
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, object := range page.Contents {
			out = append(out, storage.ObjectInfo{
				Key:          aws.ToString(object.Key),
				Size:         aws.ToInt64(object.Size),
				ETag:         strings.Trim(aws.ToString(object.ETag), "\""),
				LastModified: aws.ToTime(object.LastModified),
			})
		}
	}
	return out, nil
}

func headersFromHTTP(headers http.Header) map[string]string {
	out := map[string]string{}
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		out[key] = values[0]
	}
	return out
}

func isNotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound")
}
