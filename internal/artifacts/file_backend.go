package artifacts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type fileBackend struct {
	dir string
}

func newFileBackend(dir string) (*fileBackend, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &fileBackend{dir: dir}, nil
}

func (b *fileBackend) Root() string {
	return b.dir
}

func (b *fileBackend) Location(id string) string {
	if id == "" {
		return b.dir
	}
	return filepath.Join(b.dir, id+".md")
}

func (b *fileBackend) List(ctx context.Context) ([]storedMarkdown, error) {
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		return nil, err
	}
	var out []storedMarkdown
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-len(".md")]
		stored, err := b.Read(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, stored)
	}
	return out, nil
}

func (b *fileBackend) Read(_ context.Context, id string) (storedMarkdown, error) {
	path := b.Location(id)
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return storedMarkdown{}, errBackendNotFound
		}
		return storedMarkdown{}, err
	}
	return storedMarkdown{ID: id, Location: path, Markdown: string(raw)}, nil
}

func (b *fileBackend) Write(_ context.Context, id, markdown string) (storedMarkdown, error) {
	tmp, err := os.CreateTemp(b.dir, "."+id+"-*.tmp")
	if err != nil {
		return storedMarkdown{}, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(markdown); err != nil {
		_ = tmp.Close()
		return storedMarkdown{}, err
	}
	if err := tmp.Close(); err != nil {
		return storedMarkdown{}, err
	}
	if err := os.Rename(tmpName, b.Location(id)); err != nil {
		return storedMarkdown{}, err
	}
	return storedMarkdown{ID: id, Location: b.Location(id), Markdown: markdown}, nil
}
