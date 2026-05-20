package mcpserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotProjectSurfacesReadmeDocsAndTechStack(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Demo\n\nA useful project.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "guide.md"), []byte("# Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := SnapshotProject(root)
	if s.ReadmeTitle != "Demo" {
		t.Fatalf("readme title = %q", s.ReadmeTitle)
	}
	if len(s.Docs) != 1 || s.Docs[0] != "docs/guide.md" {
		t.Fatalf("docs = %#v", s.Docs)
	}
	if len(s.TechStack) != 1 || s.TechStack[0] != "go" {
		t.Fatalf("tech stack = %#v", s.TechStack)
	}
}
