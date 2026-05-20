package skills

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FilesystemRegistry struct{ bundles []Bundle }

func NewFilesystemRegistry(root string) (*FilesystemRegistry, error) {
	if root == "" {
		return &FilesystemRegistry{}, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return &FilesystemRegistry{}, nil
		}
		return nil, err
	}
	var bundles []Bundle
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		skillPath := filepath.Join(root, name, "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		desc := firstMarkdownLine(data)
		bundle := Bundle{Name: name, Description: desc, Version: "filesystem", Files: []File{{
			RelPath:  "SKILL.md",
			MIMEType: "text/markdown",
			Content:  staticContent(data),
		}}}
		_ = filepath.WalkDir(filepath.Join(root, name), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Base(path) == "SKILL.md" {
				return nil
			}
			rel, err := filepath.Rel(filepath.Join(root, name), path)
			if err != nil {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			bundle.Files = append(bundle.Files, File{RelPath: filepath.ToSlash(rel), MIMEType: mimeFor(rel), Content: staticContent(data)})
			return nil
		})
		bundles = append(bundles, bundle)
	}
	sort.Slice(bundles, func(i, j int) bool { return bundles[i].Name < bundles[j].Name })
	return &FilesystemRegistry{bundles: bundles}, nil
}

func (r *FilesystemRegistry) All() []Bundle { return r.bundles }
func (r *FilesystemRegistry) Get(name string) (Bundle, bool) {
	for _, b := range r.bundles {
		if b.Name == name {
			return b, true
		}
	}
	return Bundle{}, false
}
func (r *FilesystemRegistry) For(hasProject bool) []Bundle { return r.bundles }

func staticContent(data []byte) func(ProjectContext) []byte {
	copyData := append([]byte(nil), data...)
	return func(ProjectContext) []byte { return append([]byte(nil), copyData...) }
}

func firstMarkdownLine(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" && line != "---" {
			return line
		}
	}
	return "Filesystem skill"
}

func mimeFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	default:
		return "text/plain"
	}
}
