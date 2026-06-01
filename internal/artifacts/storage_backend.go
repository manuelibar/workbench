package artifacts

import (
	"context"
	"errors"
	"strings"

	"github.com/manuelibar/workbench/internal/errs"
	"github.com/manuelibar/workbench/internal/storageclient"
)

type StorageBackendOptions struct {
	Client       *storageclient.Client
	OrgID        string
	ProjectID    string
	ResourceType string
}

type storageBackend struct {
	client       *storageclient.Client
	orgID        string
	projectID    string
	resourceType string
}

func NewStorageStore(opts StorageBackendOptions, registry Registry) (*Store, error) {
	backend, err := newStorageBackend(opts)
	if err != nil {
		return nil, err
	}
	return newStoreWithBackend(backend, registry), nil
}

func newStorageBackend(opts StorageBackendOptions) (*storageBackend, error) {
	if opts.Client == nil {
		return nil, invalidBackend("storage client is required")
	}
	orgID := firstNonEmpty(opts.OrgID, "local")
	projectID := firstNonEmpty(opts.ProjectID, "workbench")
	resourceType := firstNonEmpty(opts.ResourceType, "artifacts")
	if _, err := storageclient.NewResourceRef(orgID, projectID, resourceType, "validation"); err != nil {
		return nil, err
	}
	return &storageBackend{
		client:       opts.Client,
		orgID:        orgID,
		projectID:    projectID,
		resourceType: resourceType,
	}, nil
}

func (b *storageBackend) Root() string {
	return "storage:///" + strings.Join([]string{b.orgID, b.projectID, b.resourceType}, "/")
}

func (b *storageBackend) Location(id string) string {
	if strings.TrimSpace(id) == "" {
		return b.Root()
	}
	return b.ref(id).URI()
}

func (b *storageBackend) List(ctx context.Context) ([]storedMarkdown, error) {
	summaries, err := b.client.ListResources(ctx, b.orgID, b.projectID, b.resourceType)
	if err != nil {
		return nil, err
	}
	out := make([]storedMarkdown, 0, len(summaries))
	for _, summary := range summaries {
		stored, err := b.Read(ctx, summary.Ref.ResourceID)
		if err != nil {
			continue
		}
		out = append(out, stored)
	}
	return out, nil
}

func (b *storageBackend) Read(ctx context.Context, id string) (storedMarkdown, error) {
	markdown, err := b.client.DownloadMarkdown(ctx, b.ref(id))
	if err != nil {
		if errors.Is(err, storageclient.ErrNotFound) {
			return storedMarkdown{}, errBackendNotFound
		}
		return storedMarkdown{}, err
	}
	return storedMarkdown{ID: id, Location: b.Location(id), Markdown: markdown}, nil
}

func (b *storageBackend) Write(ctx context.Context, id, markdown string) (storedMarkdown, error) {
	ref := b.ref(id)
	if err := b.client.PutMarkdown(ctx, ref, markdown); err != nil {
		return storedMarkdown{}, err
	}
	return storedMarkdown{ID: id, Location: b.Location(id), Markdown: markdown}, nil
}

func (b *storageBackend) ref(id string) storageclient.ResourceRef {
	return storageclient.ResourceRef{
		OrgID:        b.orgID,
		ProjectID:    b.projectID,
		ResourceType: b.resourceType,
		ResourceID:   strings.TrimSpace(id),
	}
}

func invalidBackend(message string) error {
	return errs.New(
		message,
		errs.WithSentinel(errs.ErrInvalid),
		errs.WithCode(CodeStoreUnavailable),
		errs.WithSeverity(errs.SeverityError),
		errs.WithRetryable(false),
	)
}
