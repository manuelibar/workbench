package mcpserver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const artifactTimeFormat = time.RFC3339

var requiredArtifactFrontmatter = []string{"id", "type", "title", "status", "created", "updated"}

type ArtifactStore struct {
	dir      string
	registry ContractRegistry
	mu       sync.Mutex
	now      func() time.Time
}

type ArtifactSummary struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Path    string `json:"path"`
}

type Artifact struct {
	ArtifactSummary
	Markdown string            `json:"markdown"`
	Sections map[string]string `json:"sections,omitempty"`
}

type ArtifactValidation struct {
	ArtifactID string   `json:"artifact_id"`
	Valid      bool     `json:"valid"`
	Issues     []string `json:"issues"`
}

type BeginArtifactRequest struct {
	Type   string `json:"type" jsonschema:"artifact contract type such as rfc, adr, charter, spec, risk, assumption"`
	Title  string `json:"title" jsonschema:"artifact title"`
	Status string `json:"status,omitempty" jsonschema:"artifact status; defaults to draft"`
	Focus  string `json:"focus,omitempty" jsonschema:"optional focus to record in the generated draft"`
	Select bool   `json:"select,omitempty" jsonschema:"select the new artifact in context after creation"`
}

type UpdateArtifactRequest struct {
	ArtifactID   string            `json:"artifact_id,omitempty" jsonschema:"artifact id; defaults to selected artifact"`
	Title        string            `json:"title,omitempty" jsonschema:"new artifact title"`
	Status       string            `json:"status,omitempty" jsonschema:"new artifact status"`
	SetSections  map[string]string `json:"set_sections,omitempty" jsonschema:"replace section bodies by section key"`
	ClearSection []string          `json:"clear_section,omitempty" jsonschema:"section keys to clear"`
}

func NewArtifactStore(dir string, registry ContractRegistry) (*ArtifactStore, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("artifact dir is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &ArtifactStore{
		dir:      dir,
		registry: registry,
		now:      time.Now,
	}, nil
}

func (s *ArtifactStore) Dir() string {
	return s.dir
}

func (s *ArtifactStore) Registry() ContractRegistry {
	return s.registry
}

func (s *ArtifactStore) Begin(req BeginArtifactRequest) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	typ := normalizeArtifactType(req.Type)
	contract, ok := s.registry.Get(typ)
	if !ok {
		return Artifact{}, fmt.Errorf("artifact.begin: unknown artifact type %q", req.Type)
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = contract.Title + " Draft"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	now := s.now().UTC().Format(artifactTimeFormat)
	id := uuid.NewString()
	artifact := Artifact{
		ArtifactSummary: ArtifactSummary{
			ID:      id,
			Type:    typ,
			Title:   title,
			Status:  status,
			Created: now,
			Updated: now,
			Path:    s.pathFor(id),
		},
		Sections: map[string]string{},
	}
	artifact.Markdown = renderArtifactMarkdown(artifact, contract, strings.TrimSpace(req.Focus))
	if err := s.write(id, artifact.Markdown); err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *ArtifactStore) List() ([]ArtifactSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []ArtifactSummary
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		artifact, err := s.readLocked(id)
		if err != nil {
			continue
		}
		out = append(out, artifact.ArtifactSummary)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Updated == out[j].Updated {
			return out[i].ID < out[j].ID
		}
		return out[i].Updated > out[j].Updated
	})
	return out, nil
}

func (s *ArtifactStore) Get(id string) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(id)
}

func (s *ArtifactStore) Exists(id string) bool {
	_, err := s.Get(id)
	return err == nil
}

func (s *ArtifactStore) Update(id string, req UpdateArtifactRequest) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	artifact, err := s.readLocked(id)
	if err != nil {
		return Artifact{}, err
	}
	contract, ok := s.registry.Get(artifact.Type)
	if !ok {
		return Artifact{}, fmt.Errorf("artifact.update: unknown artifact type %q", artifact.Type)
	}
	if title := strings.TrimSpace(req.Title); title != "" {
		artifact.Title = title
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		artifact.Status = status
	}
	if artifact.Sections == nil {
		artifact.Sections = map[string]string{}
	}
	for key, body := range req.SetSections {
		artifact.Sections[normalizeSectionKey(key)] = strings.TrimSpace(body)
	}
	for _, key := range req.ClearSection {
		artifact.Sections[normalizeSectionKey(key)] = ""
	}
	artifact.Updated = s.now().UTC().Format(artifactTimeFormat)
	artifact.Markdown = renderArtifactMarkdown(artifact, contract, "")
	if err := s.write(id, artifact.Markdown); err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *ArtifactStore) Guidance(id, typ string) (ArtifactContract, []string, error) {
	if strings.TrimSpace(id) != "" {
		artifact, err := s.Get(id)
		if err != nil {
			return ArtifactContract{}, nil, err
		}
		typ = artifact.Type
	}
	contract, ok := s.registry.Get(typ)
	if !ok {
		return ArtifactContract{}, nil, fmt.Errorf("artifact.guidance: unknown artifact type %q", typ)
	}
	next := make([]string, 0, len(contract.RequiredSections))
	for _, section := range contract.RequiredSections {
		next = append(next, fmt.Sprintf("Fill `%s`: %s", section.Key, section.Prompt))
	}
	return contract, next, nil
}

func (s *ArtifactStore) Validate(id string) (ArtifactValidation, error) {
	artifact, err := s.Get(id)
	if err != nil {
		return ArtifactValidation{}, err
	}
	validation := ArtifactValidation{ArtifactID: id}
	issues := missingFrontmatterKeys(artifact.Markdown)
	for _, key := range issues {
		validation.Issues = append(validation.Issues, "missing frontmatter key: "+key)
	}
	contract, ok := s.registry.Get(artifact.Type)
	if !ok {
		validation.Issues = append(validation.Issues, "unknown artifact type: "+artifact.Type)
		validation.Valid = false
		return validation, nil
	}
	for _, section := range contract.RequiredSections {
		body := strings.TrimSpace(artifact.Sections[section.Key])
		if sectionBodyMissing(body) {
			validation.Issues = append(validation.Issues, "missing required section body: "+section.Key)
		}
	}
	validation.Valid = len(validation.Issues) == 0
	return validation, nil
}

func (s *ArtifactStore) readLocked(id string) (Artifact, error) {
	if err := validateArtifactID(id); err != nil {
		return Artifact{}, err
	}
	path := s.pathFor(id)
	raw, err := os.ReadFile(path)
	if err != nil {
		return Artifact{}, err
	}
	artifact, err := parseArtifactMarkdown(id, path, string(raw))
	if err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *ArtifactStore) write(id, markdown string) error {
	if err := validateArtifactID(id); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, "."+id+"-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(markdown); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.pathFor(id))
}

func (s *ArtifactStore) pathFor(id string) string {
	return filepath.Join(s.dir, id+".md")
}

func validateArtifactID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("artifact id is required")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid artifact id %q", id)
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("invalid artifact id %q", id)
	}
	return nil
}

func renderArtifactMarkdown(artifact Artifact, contract ArtifactContract, focus string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: " + yamlString(artifact.ID) + "\n")
	b.WriteString("type: " + yamlString(artifact.Type) + "\n")
	b.WriteString("title: " + yamlString(artifact.Title) + "\n")
	b.WriteString("status: " + yamlString(artifact.Status) + "\n")
	b.WriteString("created: " + yamlString(artifact.Created) + "\n")
	b.WriteString("updated: " + yamlString(artifact.Updated) + "\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + artifact.Title + "\n\n")
	if strings.TrimSpace(focus) != "" {
		b.WriteString("Focus: " + strings.TrimSpace(focus) + "\n\n")
	}
	writeSections := func(sections []ArtifactSectionSpec) {
		for _, section := range sections {
			b.WriteString("## " + section.Title + "\n\n")
			body := strings.TrimSpace(artifact.Sections[section.Key])
			if body != "" {
				b.WriteString(body + "\n")
			}
			b.WriteString("\n")
		}
	}
	writeSections(contract.RequiredSections)
	writeSections(contract.OptionalSections)
	known := map[string]bool{}
	for _, section := range contract.RequiredSections {
		known[section.Key] = true
	}
	for _, section := range contract.OptionalSections {
		known[section.Key] = true
	}
	var extras []string
	for key := range artifact.Sections {
		if !known[key] {
			extras = append(extras, key)
		}
	}
	sort.Strings(extras)
	for _, key := range extras {
		b.WriteString("## " + titleFromSectionKey(key) + "\n\n")
		if body := strings.TrimSpace(artifact.Sections[key]); body != "" {
			b.WriteString(body + "\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func parseArtifactMarkdown(id, path, markdown string) (Artifact, error) {
	frontmatter, body, err := parseFrontmatter(markdown)
	if err != nil {
		return Artifact{}, err
	}
	summary := ArtifactSummary{
		ID:      firstNonEmpty(frontmatter["id"], id),
		Type:    normalizeArtifactType(frontmatter["type"]),
		Title:   frontmatter["title"],
		Status:  firstNonEmpty(frontmatter["status"], "draft"),
		Created: frontmatter["created"],
		Updated: frontmatter["updated"],
		Path:    path,
	}
	if summary.Title == "" {
		summary.Title = summary.ID
	}
	return Artifact{
		ArtifactSummary: summary,
		Markdown:        markdown,
		Sections:        parseMarkdownSections(body),
	}, nil
}

func parseFrontmatter(markdown string) (map[string]string, string, error) {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	if !strings.HasPrefix(markdown, "---\n") {
		return map[string]string{}, markdown, errors.New("artifact: missing frontmatter")
	}
	rest := strings.TrimPrefix(markdown, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return map[string]string{}, markdown, errors.New("artifact: unclosed frontmatter")
	}
	rawFront := rest[:idx]
	body := rest[idx+len("\n---\n"):]
	out := map[string]string{}
	for _, line := range strings.Split(rawFront, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		out[key] = value
	}
	return out, body, nil
}

func parseMarkdownSections(body string) map[string]string {
	sections := map[string]string{}
	var current string
	var b strings.Builder
	flush := func() {
		if current != "" {
			sections[current] = strings.TrimSpace(b.String())
			b.Reset()
		}
	}
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		if title, ok := strings.CutPrefix(line, "## "); ok {
			flush()
			current = normalizeSectionKey(title)
			continue
		}
		if current != "" {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	flush()
	return sections
}

func missingFrontmatterKeys(markdown string) []string {
	frontmatter, _, err := parseFrontmatter(markdown)
	if err != nil {
		return append([]string(nil), requiredArtifactFrontmatter...)
	}
	var missing []string
	for _, key := range requiredArtifactFrontmatter {
		if strings.TrimSpace(frontmatter[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func sectionBodyMissing(body string) bool {
	switch strings.ToLower(strings.TrimSpace(body)) {
	case "", "todo", "tbd", "n/a":
		return true
	default:
		return false
	}
}

func yamlString(v string) string {
	return strconv.Quote(v)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
