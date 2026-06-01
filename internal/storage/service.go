package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	metadataOrgID          = "org_id"
	metadataProjectID      = "project_id"
	metadataResourceType   = "resource_type"
	metadataResourceID     = "resource_id"
	metadataSourceMIMEType = "source_mime_type"
	metadataIndexed        = "storage_indexed"
)

type Object struct {
	Key         string
	Body        []byte
	ContentType string
	Metadata    map[string]string
}

type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	ContentType  string
	Metadata     map[string]string
}

type SignedURL struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt string            `json:"expires_at,omitempty"`
}

type MultipartPart struct {
	PartNumber int               `json:"part_number"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type MultipartUpload struct {
	Key      string          `json:"key"`
	UploadID string          `json:"upload_id"`
	Parts    []MultipartPart `json:"parts"`
}

type ObjectStore interface {
	PresignPut(context.Context, string, string, map[string]string, time.Duration) (SignedURL, error)
	PresignGet(context.Context, string, time.Duration) (SignedURL, error)
	StartMultipart(context.Context, string, string, map[string]string, int, time.Duration) (MultipartUpload, error)
	GetObject(context.Context, string, string) (Object, error)
	PutObject(context.Context, string, []byte, string, map[string]string) error
	DeleteObject(context.Context, string) error
	HeadObject(context.Context, string) (ObjectInfo, error)
	ListObjects(context.Context, string) ([]ObjectInfo, error)
}

type Service struct {
	objects        ObjectStore
	normalizer     Normalizer
	presignTTL     time.Duration
	now            func() time.Time
	supportedMIMEs map[string]bool
}

type ServiceOptions struct {
	Objects        ObjectStore
	Normalizer     Normalizer
	PresignTTL     time.Duration
	SupportedMIMEs []string
}

type UploadURLRequest struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id,omitempty"`
	Filename     string `json:"filename"`
	MIMEType     string `json:"mime_type"`
}

type UploadURLResult struct {
	Ref       ResourceRef `json:"resource"`
	Key       string      `json:"key"`
	UploadURL SignedURL   `json:"upload_url"`
}

type DownloadURLResult struct {
	Ref           ResourceRef `json:"resource"`
	Key           string      `json:"key"`
	DownloadURL   SignedURL   `json:"download_url"`
	SupportsRange bool        `json:"supports_range"`
}

type UpdateURLRequest struct {
	ResourceRef
	MIMEType string `json:"mime_type"`
}

type UpdateURLResult struct {
	Ref       ResourceRef `json:"resource"`
	Key       string      `json:"key"`
	UploadURL SignedURL   `json:"upload_url"`
}

type MultipartStartRequest struct {
	ResourceRef
	Filename  string `json:"filename,omitempty"`
	MIMEType  string `json:"mime_type"`
	PartCount int    `json:"part_count"`
}

type MultipartStartResult struct {
	Ref    ResourceRef     `json:"resource"`
	Upload MultipartUpload `json:"upload"`
}

type ResourceSummary struct {
	Ref          ResourceRef `json:"resource"`
	Key          string      `json:"key"`
	URI          string      `json:"uri"`
	Size         int64       `json:"size"`
	ETag         string      `json:"etag,omitempty"`
	LastModified string      `json:"last_modified,omitempty"`
}

func NewService(opts ServiceOptions) (*Service, error) {
	if opts.Objects == nil {
		return nil, invalid("storage.objects.required", "object store is required")
	}
	normalizer := opts.Normalizer
	if normalizer == nil {
		normalizer = MarkItDownNormalizer{}
	}
	ttl := opts.PresignTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	mimes := defaultSupportedMIMEs()
	for _, mimeType := range opts.SupportedMIMEs {
		if normalized := normalizeMIMEType(mimeType); normalized != "" {
			mimes[normalized] = true
		}
	}
	return &Service{
		objects:        opts.Objects,
		normalizer:     normalizer,
		presignTTL:     ttl,
		now:            time.Now,
		supportedMIMEs: mimes,
	}, nil
}

func (s *Service) UploadURL(ctx context.Context, req UploadURLRequest) (UploadURLResult, error) {
	if err := s.validateMIME(req.MIMEType); err != nil {
		return UploadURLResult{}, err
	}
	resourceID := strings.TrimSpace(req.ResourceID)
	if resourceID == "" {
		resourceID = uuid.NewString()
	}
	ref, err := NewResourceRef(req.OrgID, req.ProjectID, req.ResourceType, resourceID)
	if err != nil {
		return UploadURLResult{}, err
	}
	key, err := InboxKey(ref.OrgID, ref.ProjectID, req.Filename)
	if err != nil {
		return UploadURLResult{}, err
	}
	metadata := metadataFor(ref, req.MIMEType, false)
	signed, err := s.objects.PresignPut(ctx, key, normalizeMIMEType(req.MIMEType), metadata, s.presignTTL)
	if err != nil {
		return UploadURLResult{}, dependency("storage.presign.put", "Upload URL generation failed", err)
	}
	return UploadURLResult{Ref: ref, Key: key, UploadURL: signed}, nil
}

func (s *Service) DownloadURL(ctx context.Context, ref ResourceRef) (DownloadURLResult, error) {
	if err := ref.Validate(); err != nil {
		return DownloadURLResult{}, err
	}
	key := ref.Key()
	signed, err := s.objects.PresignGet(ctx, key, s.presignTTL)
	if err != nil {
		return DownloadURLResult{}, dependency("storage.presign.get", "Download URL generation failed", err)
	}
	return DownloadURLResult{Ref: ref, Key: key, DownloadURL: signed, SupportsRange: true}, nil
}

func (s *Service) UpdateURL(ctx context.Context, req UpdateURLRequest) (UpdateURLResult, error) {
	if err := req.ResourceRef.Validate(); err != nil {
		return UpdateURLResult{}, err
	}
	if err := s.validateMIME(req.MIMEType); err != nil {
		return UpdateURLResult{}, err
	}
	key := req.ResourceRef.Key()
	metadata := metadataFor(req.ResourceRef, req.MIMEType, false)
	signed, err := s.objects.PresignPut(ctx, key, normalizeMIMEType(req.MIMEType), metadata, s.presignTTL)
	if err != nil {
		return UpdateURLResult{}, dependency("storage.presign.put", "Update URL generation failed", err)
	}
	return UpdateURLResult{Ref: req.ResourceRef, Key: key, UploadURL: signed}, nil
}

func (s *Service) StartMultipart(ctx context.Context, req MultipartStartRequest) (MultipartStartResult, error) {
	if err := req.ResourceRef.Validate(); err != nil {
		return MultipartStartResult{}, err
	}
	if req.PartCount <= 0 {
		return MultipartStartResult{}, invalid("storage.multipart.part_count", "part_count must be positive")
	}
	if err := s.validateMIME(req.MIMEType); err != nil {
		return MultipartStartResult{}, err
	}
	metadata := metadataFor(req.ResourceRef, req.MIMEType, false)
	upload, err := s.objects.StartMultipart(ctx, req.ResourceRef.Key(), normalizeMIMEType(req.MIMEType), metadata, req.PartCount, s.presignTTL)
	if err != nil {
		return MultipartStartResult{}, dependency("storage.multipart.start", "Multipart upload start failed", err)
	}
	return MultipartStartResult{Ref: req.ResourceRef, Upload: upload}, nil
}

func (s *Service) Stats(ctx context.Context, ref ResourceRef) (ResourceStats, error) {
	if err := ref.Validate(); err != nil {
		return ResourceStats{}, err
	}
	object, err := s.objects.GetObject(ctx, ref.Key(), "bytes=0-8192")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ResourceStats{}, err
		}
		return ResourceStats{}, dependency("storage.stats.get", "Storage stats read failed", err)
	}
	stats, err := ParseStats(string(object.Body))
	if err != nil {
		return ResourceStats{}, err
	}
	return stats, nil
}

func (s *Service) List(ctx context.Context, orgID, projectID, resourceType string) ([]ResourceSummary, error) {
	prefix, err := Prefix(orgID, projectID, resourceType)
	if err != nil {
		return nil, err
	}
	objects, err := s.objects.ListObjects(ctx, prefix)
	if err != nil {
		return nil, dependency("storage.list", "Storage resource list failed", err)
	}
	out := make([]ResourceSummary, 0, len(objects))
	for _, object := range objects {
		ref, ok := RefFromKey(object.Key)
		if !ok {
			continue
		}
		summary := ResourceSummary{
			Ref:  ref,
			Key:  object.Key,
			URI:  ref.URI(),
			Size: object.Size,
			ETag: object.ETag,
		}
		if !object.LastModified.IsZero() {
			summary.LastModified = object.LastModified.UTC().Format(time.RFC3339)
		}
		out = append(out, summary)
	}
	return out, nil
}

func (s *Service) HandleS3Event(ctx context.Context, body io.Reader) error {
	var event s3Event
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		return invalid("storage.s3_event.invalid", "S3 event payload is invalid")
	}
	if len(event.Records) == 0 {
		return invalid("storage.s3_event.empty", "S3 event payload contains no records")
	}
	for _, record := range event.Records {
		key := strings.TrimLeft(record.S3.Object.Key, "/")
		if decoded, err := url.QueryUnescape(key); err == nil {
			key = decoded
		}
		if key == "" {
			return invalid("storage.s3_event.key_required", "S3 event record object key is required")
		}
		if err := s.ProcessObject(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ProcessObject(ctx context.Context, key string) error {
	key = strings.Trim(strings.TrimSpace(key), "/")
	if key == "" {
		return invalid("storage.object.key_required", "object key is required")
	}
	info, err := s.objects.HeadObject(ctx, key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return err
		}
		return dependency("storage.object.head", "Storage object metadata read failed", err)
	}
	if isIndexed(info.Metadata) {
		return nil
	}
	ref, targetKey, err := refForObject(key, info.Metadata)
	if err != nil {
		return err
	}
	sourceMIMEType := firstNonEmpty(metadataValue(info.Metadata, metadataSourceMIMEType), info.ContentType, "application/octet-stream")
	if err := s.validateMIME(sourceMIMEType); err != nil {
		return err
	}
	object, err := s.objects.GetObject(ctx, key, "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return err
		}
		return dependency("storage.object.get", "Storage object read failed", err)
	}
	markdown, err := s.normalizer.Normalize(ctx, NormalizationInput{
		Filename: path.Base(key),
		MIMEType: sourceMIMEType,
		Data:     object.Body,
	})
	if err != nil {
		return err
	}
	indexed, err := IndexMarkdown(ref, markdown, sourceMIMEType, s.now())
	if err != nil {
		return err
	}
	if err := validateIndexOffsets(indexed.Markdown, indexed.Stats.Index.Sections); err != nil {
		return dependency("storage.index.invalid", "Storage index generation failed", err)
	}
	if err := s.objects.PutObject(ctx, targetKey, []byte(indexed.Markdown), "text/markdown", metadataFor(ref, sourceMIMEType, true)); err != nil {
		return dependency("storage.object.put", "Indexed Markdown write failed", err)
	}
	if strings.HasPrefix(key, "inbox/") {
		if err := s.objects.DeleteObject(ctx, key); err != nil {
			return dependency("storage.object.delete", "Inbox object delete failed", err)
		}
	}
	return nil
}

func metadataFor(ref ResourceRef, mimeType string, indexed bool) map[string]string {
	return map[string]string{
		metadataOrgID:          ref.OrgID,
		metadataProjectID:      ref.ProjectID,
		metadataResourceType:   ref.ResourceType,
		metadataResourceID:     ref.ResourceID,
		metadataSourceMIMEType: normalizeMIMEType(mimeType),
		metadataIndexed:        strconv.FormatBool(indexed),
	}
}

func refForObject(key string, metadata map[string]string) (ResourceRef, string, error) {
	if ref, ok := RefFromKey(key); ok {
		return ref, ref.Key(), nil
	}
	if !strings.HasPrefix(key, "inbox/") {
		return ResourceRef{}, "", invalid("storage.object.key_invalid", "object key is not a storage resource or inbox key")
	}
	ref, err := NewResourceRef(
		metadataValue(metadata, metadataOrgID),
		metadataValue(metadata, metadataProjectID),
		metadataValue(metadata, metadataResourceType),
		metadataValue(metadata, metadataResourceID),
	)
	if err != nil {
		return ResourceRef{}, "", err
	}
	return ref, ref.Key(), nil
}

func metadataValue(metadata map[string]string, key string) string {
	for k, value := range metadata {
		if strings.EqualFold(strings.ReplaceAll(k, "-", "_"), key) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isIndexed(metadata map[string]string) bool {
	indexed, _ := strconv.ParseBool(metadataValue(metadata, metadataIndexed))
	return indexed
}

func (s *Service) validateMIME(mimeType string) error {
	mimeType = normalizeMIMEType(mimeType)
	if mimeType == "" {
		return invalid("storage.mime.required", "mime_type is required")
	}
	if !s.supportedMIMEs[mimeType] {
		return invalid("storage.mime.unsupported", "mime_type %q is not supported", mimeType)
	}
	return nil
}

func normalizeMIMEType(mimeType string) string {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if base, _, ok := strings.Cut(mimeType, ";"); ok {
		mimeType = strings.TrimSpace(base)
	}
	return mimeType
}

func defaultSupportedMIMEs() map[string]bool {
	values := []string{
		"text/markdown",
		"text/plain",
		"text/csv",
		"text/html",
		"application/json",
		"application/pdf",
		"application/msword",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"image/png",
		"image/jpeg",
		"image/gif",
		"audio/mpeg",
		"audio/wav",
		"audio/x-wav",
	}
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type s3Event struct {
	Records []struct {
		S3 struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}
