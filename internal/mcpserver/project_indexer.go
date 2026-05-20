package mcpserver

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ProjectSnapshot struct {
	Root        string   `json:"root"`
	ReadmePath  string   `json:"readme_path,omitempty"`
	ReadmeTitle string   `json:"readme_title,omitempty"`
	ReadmeBrief string   `json:"readme_brief,omitempty"`
	Docs        []string `json:"docs"`
	TechStack   []string `json:"tech_stack"`
}

func SnapshotProject(root string) ProjectSnapshot {
	if root == "" {
		root, _ = os.Getwd()
	}
	root, _ = filepath.Abs(root)
	s := ProjectSnapshot{Root: root, Docs: []string{}, TechStack: DetectTechStack(root)}
	for _, name := range []string{"README.md", "readme.md", "Readme.md"} {
		path := filepath.Join(root, name)
		if data, err := os.ReadFile(path); err == nil {
			s.ReadmePath = path
			s.ReadmeTitle, s.ReadmeBrief = summarizeMarkdown(data)
			break
		}
	}
	docsRoot := filepath.Join(root, "docs")
	_ = filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err == nil {
			s.Docs = append(s.Docs, rel)
		}
		return nil
	})
	sort.Strings(s.Docs)
	return s
}

func DetectTechStack(root string) []string {
	checks := []struct{ file, tech string }{
		{"go.mod", "go"},
		{"package.json", "nodejs"},
		{"pyproject.toml", "python"},
		{"Cargo.toml", "rust"},
		{"Dockerfile", "docker"},
		{"docker-compose.yml", "docker-compose"},
	}
	out := []string{}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			out = append(out, c.tech)
		}
	}
	return out
}

func summarizeMarkdown(data []byte) (title, brief string) {
	lines := strings.Split(string(data), "\n")
	paras := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if title == "" && strings.HasPrefix(line, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			continue
		}
		paras = append(paras, line)
		if len(strings.Join(paras, " ")) > 420 {
			break
		}
	}
	brief = strings.Join(paras, " ")
	if len(brief) > 500 {
		brief = brief[:500]
	}
	return title, brief
}
