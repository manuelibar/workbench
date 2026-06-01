package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/manuelibar/workbench/internal/errs"
)

func TestResolveArtifactIDClassifiesMissingSelection(t *testing.T) {
	server, err := New(Options{ArtifactDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := server.resolveArtifactID(context.Background(), ""); err == nil {
		t.Fatal("missing selection returned nil error")
	} else if !errors.Is(err, errs.ErrInvalid) {
		t.Fatalf("missing selection error = %v, want ErrInvalid", err)
	} else if got := errs.CodeOf(err); got != errCodeArtifactSelectionMissing {
		t.Fatalf("missing selection code = %q", got)
	}
}
