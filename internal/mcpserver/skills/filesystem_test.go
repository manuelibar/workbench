package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesystemRegistryLoadsSkillBundles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "go-guidelines")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Go Guidelines\n\nUse gofmt.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg, err := NewFilesystemRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := reg.Get("go-guidelines")
	if !ok {
		t.Fatal("bundle not loaded")
	}
	if got := string(b.Files[0].Content(ProjectContext{})); !strings.Contains(got, "gofmt") {
		t.Fatalf("content = %q", got)
	}
}
