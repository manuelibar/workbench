package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/manuelibar/workbench/internal/errs"
)

func (s *Server) materializeScopedArtifact(ctx context.Context, id string) (string, error) {
	attrs := map[string]any{
		"operation":   "artifact.scope",
		"artifact_id": id,
	}
	artifact, err := s.artifacts.GetContext(ctx, id)
	if err != nil {
		return "", errs.Decorate(err, errs.WithAttrs(attrs))
	}
	state := s.scope.Snapshot()
	if state.ArtifactID != nil && *state.ArtifactID == id &&
		state.LocalArtifactPath != nil && *state.LocalArtifactPath != "" {
		if _, err := os.Stat(*state.LocalArtifactPath); err == nil {
			return *state.LocalArtifactPath, nil
		}
	}
	path := s.scopedArtifactPath(id)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		attrs["local_artifact_path"] = path
		return "", errs.New(
			"Artifact scope required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeArtifactScopeMissing),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	if err := os.WriteFile(path, []byte(artifact.Markdown), 0o644); err != nil {
		attrs["local_artifact_path"] = path
		return "", errs.New(
			"Artifact scope required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeArtifactScopeMissing),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	return path, nil
}

func (s *Server) readScopedArtifactMarkdown(ctx context.Context, id string) (string, error) {
	state := s.scope.Snapshot()
	if state.ArtifactID != nil && *state.ArtifactID == id &&
		state.LocalArtifactPath != nil && *state.LocalArtifactPath != "" {
		if raw, err := os.ReadFile(*state.LocalArtifactPath); err == nil {
			return string(raw), nil
		}
	}
	artifact, err := s.artifacts.GetContext(ctx, id)
	if err != nil {
		return "", err
	}
	return artifact.Markdown, nil
}

func (s *Server) scopedArtifact() (string, string, error) {
	state := s.scope.Snapshot()
	if state.ArtifactID == nil || strings.TrimSpace(*state.ArtifactID) == "" {
		return "", "", errs.New(
			"Artifact scope required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(errCodeArtifactScopeMissing),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(map[string]any{"operation": "artifact.scope.require"}),
			errs.WithRetryable(false),
		)
	}
	path := ""
	if state.LocalArtifactPath != nil {
		path = strings.TrimSpace(*state.LocalArtifactPath)
	}
	return strings.TrimSpace(*state.ArtifactID), path, nil
}

func (s *Server) scopedArtifactPath(id string) string {
	return filepath.Join(s.scopeDir, id+".md")
}
