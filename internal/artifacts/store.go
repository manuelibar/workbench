package artifacts

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/manuelibar/workbench/internal/errs"
)

const (
	artifactTimeFormat = time.RFC3339

	CodeDirInvalid       errs.Code = "workbench.artifact.dir_invalid"
	CodeInvalid          errs.Code = "workbench.artifact.invalid"
	CodeIDInvalid        errs.Code = "workbench.artifact.id_invalid"
	CodeFindFailed       errs.Code = "workbench.artifact.find_failed"
	CodeNotFound         errs.Code = "workbench.artifact.not_found"
	CodeReadFailed       errs.Code = "workbench.artifact.read_failed"
	CodeStoreUnavailable errs.Code = "workbench.artifact.store_unavailable"
	CodeTypeUnknown      errs.Code = "workbench.artifact.type_unknown"
	CodeWriteFailed      errs.Code = "workbench.artifact.write_failed"
)

var requiredArtifactFrontmatter = []string{"id", "type", "title", "status", "created", "updated"}

type Store struct {
	backend  markdownBackend
	registry Registry
	mu       sync.Mutex
	now      func() time.Time
}

type Summary struct {
	ID      string
	Type    string
	Title   string
	Status  string
	Created string
	Updated string
	Path    string
}

type Artifact struct {
	Summary
	Markdown string
	Sections map[string]string
}

type Validation struct {
	ArtifactID string
	Valid      bool
	Issues     []string
}

type CreateRequest struct {
	Type   string
	Title  string
	Status string
	Focus  string
}

type UpdateRequest struct {
	Title        string
	Status       string
	SetSections  map[string]string
	ClearSection []string
}

func NewStore(dir string, registry Registry) (*Store, error) {
	attrs := map[string]any{
		"operation": "artifact.store.init",
		"path":      dir,
	}
	if strings.TrimSpace(dir) == "" {
		return nil, errs.New(
			"Artifact directory is required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeDirInvalid),
			errs.WithSeverity(errs.SeverityError),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	backend, err := newFileBackend(dir)
	if err != nil {
		attrs["operation"] = "artifact.store.mkdir"
		return nil, errs.New(
			"Artifact store unavailable",
			errs.WithSentinel(errs.ErrDependencyFailed),
			errs.WithCode(CodeStoreUnavailable),
			errs.WithSeverity(errs.SeverityError),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	return newStoreWithBackend(backend, registry), nil
}

func newStoreWithBackend(backend markdownBackend, registry Registry) *Store {
	return &Store{
		backend:  backend,
		registry: registry,
		now:      time.Now,
	}
}

func (s *Store) Dir() string {
	return s.backend.Root()
}

func (s *Store) Registry() Registry {
	return s.registry
}

func (s *Store) Create(req CreateRequest) (Artifact, error) {
	return s.CreateContext(context.Background(), req)
}

func (s *Store) CreateContext(ctx context.Context, req CreateRequest) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := map[string]any{
		"operation":     "artifact.create",
		"artifact_type": req.Type,
	}
	typ := normalizeArtifactType(req.Type)
	contract, ok := s.registry.Get(typ)
	if !ok {
		return Artifact{}, unknownTypeError("artifact.create", req.Type)
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
		Summary: Summary{
			ID:      id,
			Type:    typ,
			Title:   title,
			Status:  status,
			Created: now,
			Updated: now,
			Path:    s.backend.Location(id),
		},
		Sections: map[string]string{},
	}
	artifact.Markdown = renderArtifactMarkdown(artifact, contract, strings.TrimSpace(req.Focus))
	if err := s.write(ctx, id, artifact.Markdown); err != nil {
		attrs["artifact_id"] = id
		err = errs.Decorate(err, errs.WithAttrs(attrs))
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *Store) List() ([]Summary, error) {
	return s.ListContext(context.Background())
}

func (s *Store) ListContext(ctx context.Context) ([]Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := map[string]any{
		"operation": "artifact.store.list",
		"path":      s.backend.Root(),
	}
	stored, err := s.backend.List(ctx)
	if err != nil {
		return nil, errs.New(
			"Artifact find failed",
			errs.WithSentinel(errs.ErrDependencyFailed),
			errs.WithCode(CodeFindFailed),
			errs.WithSeverity(errs.SeverityError),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	var out []Summary
	for _, item := range stored {
		artifact, err := parseArtifactMarkdown(item.ID, item.Location, item.Markdown)
		if err != nil {
			continue
		}
		out = append(out, artifact.Summary)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Updated == out[j].Updated {
			return out[i].ID < out[j].ID
		}
		return out[i].Updated > out[j].Updated
	})
	return out, nil
}

func (s *Store) Get(id string) (Artifact, error) {
	return s.GetContext(context.Background(), id)
}

func (s *Store) GetContext(ctx context.Context, id string) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(ctx, id)
}

func (s *Store) Exists(id string) bool {
	_, err := s.Get(id)
	return err == nil
}

func (s *Store) ExistsContext(ctx context.Context, id string) bool {
	_, err := s.GetContext(ctx, id)
	return err == nil
}

func (s *Store) CheckExists(id string) error {
	_, err := s.Get(id)
	return err
}

func (s *Store) CheckExistsContext(ctx context.Context, id string) error {
	_, err := s.GetContext(ctx, id)
	return err
}

func (s *Store) Update(id string, req UpdateRequest) (Artifact, error) {
	return s.UpdateContext(context.Background(), id, req)
}

func (s *Store) UpdateContext(ctx context.Context, id string, req UpdateRequest) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := map[string]any{
		"operation":   "artifact.store.update",
		"artifact_id": id,
	}
	artifact, err := s.readLocked(ctx, id)
	if err != nil {
		return Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	contract, ok := s.registry.Get(artifact.Type)
	if !ok {
		return Artifact{}, unknownTypeError("artifact.store.update", artifact.Type)
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
	if err := s.write(ctx, id, artifact.Markdown); err != nil {
		err = errs.Decorate(err, errs.WithAttrs(attrs))
		return Artifact{}, err
	}
	return artifact, nil
}

func (s *Store) UploadMarkdown(id, markdown string) (Artifact, error) {
	return s.UploadMarkdownContext(context.Background(), id, markdown)
}

func (s *Store) UploadMarkdownContext(ctx context.Context, id, markdown string) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := map[string]any{
		"operation":   "artifact.upload",
		"artifact_id": id,
	}
	if err := validateArtifactID(id); err != nil {
		return Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	artifact, err := parseArtifactMarkdown(id, s.backend.Location(id), markdown)
	if err != nil {
		return Artifact{}, errs.New(
			"Artifact file is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	if artifact.ID != id {
		attrs["frontmatter_id"] = artifact.ID
		return Artifact{}, errs.New(
			"Artifact file is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	if missing := missingFrontmatterKeys(markdown); len(missing) > 0 {
		attrs["missing_frontmatter"] = strings.Join(missing, ",")
		return Artifact{}, errs.New(
			"Artifact file is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	if _, ok := s.registry.Get(artifact.Type); !ok {
		return Artifact{}, unknownTypeError("artifact.upload", artifact.Type)
	}
	if err := s.write(ctx, id, markdown); err != nil {
		return Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	return artifact, nil
}

func (s *Store) Guidance(id, typ string) (Contract, []string, error) {
	return s.GuidanceContext(context.Background(), id, typ)
}

func (s *Store) GuidanceContext(ctx context.Context, id, typ string) (Contract, []string, error) {
	attrs := map[string]any{
		"operation":     "artifact.store.guidance",
		"artifact_id":   id,
		"artifact_type": typ,
	}
	if strings.TrimSpace(id) != "" {
		artifact, err := s.GetContext(ctx, id)
		if err != nil {
			return Contract{}, nil, errs.Decorate(err, errs.WithAttrs(attrs))
		}
		typ = artifact.Type
		attrs["artifact_type"] = typ
	}
	contract, ok := s.registry.Get(typ)
	if !ok {
		return Contract{}, nil, unknownTypeError("artifact.store.guidance", typ)
	}
	next := make([]string, 0, len(contract.RequiredSections))
	for _, section := range contract.RequiredSections {
		next = append(next, fmt.Sprintf("Fill `%s`: %s", section.Key, section.Prompt))
	}
	return contract, next, nil
}

func (s *Store) Validate(id string) (Validation, error) {
	return s.ValidateContext(context.Background(), id)
}

func (s *Store) ValidateContext(ctx context.Context, id string) (Validation, error) {
	artifact, err := s.GetContext(ctx, id)
	if err != nil {
		return Validation{}, err
	}
	validation := Validation{ArtifactID: id}
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

func (s *Store) readLocked(ctx context.Context, id string) (Artifact, error) {
	attrs := map[string]any{
		"operation":   "artifact.read",
		"artifact_id": id,
	}
	if err := validateArtifactID(id); err != nil {
		return Artifact{}, errs.Decorate(err, errs.WithAttrs(attrs))
	}
	location := s.backend.Location(id)
	attrs["path"] = location
	stored, err := s.backend.Read(ctx, id)
	if err != nil {
		if errors.Is(err, errBackendNotFound) {
			return Artifact{}, errs.New(
				"Artifact not found",
				errs.WithSentinel(errs.ErrNotFound),
				errs.WithCode(CodeNotFound),
				errs.WithSeverity(errs.SeverityWarning),
				errs.WithCause(err),
				errs.WithAttrs(attrs),
				errs.WithRetryable(false),
			)
		}
		return Artifact{}, errs.New(
			"Artifact read failed",
			errs.WithSentinel(errs.ErrDependencyFailed),
			errs.WithCode(CodeReadFailed),
			errs.WithSeverity(errs.SeverityError),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	artifact, err := parseArtifactMarkdown(id, stored.Location, stored.Markdown)
	if err != nil {
		attrs["operation"] = "artifact.parse"
		return Artifact{}, errs.New(
			"Artifact file is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeInvalid),
			errs.WithSeverity(errs.SeverityError),
			errs.WithCause(err),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	return artifact, nil
}

func (s *Store) write(ctx context.Context, id, markdown string) error {
	attrs := map[string]any{
		"operation":   "artifact.write",
		"artifact_id": id,
	}
	if err := validateArtifactID(id); err != nil {
		return errs.Decorate(err, errs.WithAttrs(attrs))
	}
	if _, err := s.backend.Write(ctx, id, markdown); err != nil {
		return writeError(id, "artifact.write.backend", err)
	}
	return nil
}

func validateArtifactID(id string) error {
	attrs := map[string]any{"artifact_id": id}
	if strings.TrimSpace(id) == "" {
		return errs.New(
			"Artifact id is required",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeIDInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return errs.New(
			"Artifact id is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeIDInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	if strings.Contains(id, "..") {
		return errs.New(
			"Artifact id is invalid",
			errs.WithSentinel(errs.ErrInvalid),
			errs.WithCode(CodeIDInvalid),
			errs.WithSeverity(errs.SeverityWarning),
			errs.WithAttrs(attrs),
			errs.WithRetryable(false),
		)
	}
	return nil
}

func renderArtifactMarkdown(artifact Artifact, contract Contract, focus string) string {
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
	writeSections := func(sections []SectionSpec) {
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
	summary := Summary{
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
		Summary:  summary,
		Markdown: markdown,
		Sections: parseMarkdownSections(body),
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

func unknownTypeError(operation, typ string) error {
	attrs := map[string]any{
		"operation":     operation,
		"artifact_type": typ,
	}
	return errs.New(
		"Unknown artifact type",
		errs.WithSentinel(errs.ErrInvalid),
		errs.WithCode(CodeTypeUnknown),
		errs.WithSeverity(errs.SeverityWarning),
		errs.WithAttrs(attrs),
		errs.WithRetryable(false),
	)
}

func writeError(id, operation string, cause error) error {
	attrs := map[string]any{
		"operation":   operation,
		"artifact_id": id,
	}
	return errs.New(
		"Artifact write failed",
		errs.WithSentinel(errs.ErrDependencyFailed),
		errs.WithCode(CodeWriteFailed),
		errs.WithSeverity(errs.SeverityError),
		errs.WithCause(cause),
		errs.WithAttrs(attrs),
		errs.WithRetryable(false),
	)
}
