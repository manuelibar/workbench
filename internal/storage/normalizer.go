package storage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NormalizationInput struct {
	Filename string
	MIMEType string
	Data     []byte
}

type Normalizer interface {
	Normalize(context.Context, NormalizationInput) (string, error)
}

type MarkItDownNormalizer struct {
	Command string
	TempDir string
}

func (n MarkItDownNormalizer) Normalize(ctx context.Context, input NormalizationInput) (string, error) {
	command := strings.TrimSpace(n.Command)
	if command == "" {
		command = "markitdown"
	}
	name := filepath.Base(strings.TrimSpace(input.Filename))
	if name == "." || name == string(filepath.Separator) {
		name = "resource"
	}
	tmp, err := os.CreateTemp(n.TempDir, "markitdown-*-"+name)
	if err != nil {
		return "", dependency("storage.markitdown.temp", "MarkItDown normalization failed", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(input.Data); err != nil {
		_ = tmp.Close()
		return "", dependency("storage.markitdown.write", "MarkItDown normalization failed", err)
	}
	if err := tmp.Close(); err != nil {
		return "", dependency("storage.markitdown.close", "MarkItDown normalization failed", err)
	}
	cmd := exec.CommandContext(ctx, command, tmpName)
	out, err := cmd.Output()
	if err != nil {
		return "", dependency("storage.markitdown.exec", "MarkItDown normalization failed", err)
	}
	return string(out), nil
}

type PassthroughNormalizer struct{}

func (PassthroughNormalizer) Normalize(_ context.Context, input NormalizationInput) (string, error) {
	return string(input.Data), nil
}
