package storage

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type IndexedResource struct {
	Ref      ResourceRef
	Markdown string
	Stats    ResourceStats
}

type ResourceStats struct {
	Resource ResourceMetadata `json:"resource" yaml:"resource"`
	Index    ResourceIndex    `json:"index" yaml:"index"`
}

type ResourceMetadata struct {
	OrgID          string `json:"org_id" yaml:"org_id"`
	ProjectID      string `json:"project_id" yaml:"project_id"`
	ResourceType   string `json:"resource_type" yaml:"resource_type"`
	ResourceID     string `json:"resource_id" yaml:"resource_id"`
	SourceMIMEType string `json:"source_mime_type" yaml:"source_mime_type"`
	IndexedAt      string `json:"indexed_at" yaml:"indexed_at"`
	ByteLength     int    `json:"byte_length" yaml:"byte_length"`
}

type ResourceIndex struct {
	Sections []SectionIndex `json:"sections" yaml:"sections"`
}

type SectionIndex struct {
	Title     string `json:"title" yaml:"title"`
	Level     int    `json:"level" yaml:"level"`
	Anchor    string `json:"anchor" yaml:"anchor"`
	StartByte int    `json:"start_byte" yaml:"start_byte"`
	EndByte   int    `json:"end_byte" yaml:"end_byte"`
}

func IndexMarkdown(ref ResourceRef, markdown, sourceMIMEType string, now time.Time) (IndexedResource, error) {
	if err := ref.Validate(); err != nil {
		return IndexedResource{}, err
	}
	frontmatter, body, err := splitFrontmatter(markdown)
	if err != nil {
		return IndexedResource{}, err
	}
	delete(frontmatter, "storage")
	delete(frontmatter, "index")
	body = normalizeMarkdownBody(body)
	var sections []SectionIndex
	frontmatterLen := 0
	for i := 0; i < 10; i++ {
		sections = scanSections(body, frontmatterLen)
		stats := ResourceStats{
			Resource: ResourceMetadata{
				OrgID:          ref.OrgID,
				ProjectID:      ref.ProjectID,
				ResourceType:   ref.ResourceType,
				ResourceID:     ref.ResourceID,
				SourceMIMEType: strings.TrimSpace(sourceMIMEType),
				IndexedAt:      now.UTC().Format(time.RFC3339),
				ByteLength:     frontmatterLen + len(body),
			},
			Index: ResourceIndex{Sections: sections},
		}
		rendered, err := renderIndexedFrontmatter(frontmatter, stats)
		if err != nil {
			return IndexedResource{}, err
		}
		nextLen := len(rendered)
		if nextLen == frontmatterLen {
			return IndexedResource{
				Ref:      ref,
				Markdown: rendered + body,
				Stats:    stats,
			}, nil
		}
		frontmatterLen = nextLen
	}
	sections = scanSections(body, frontmatterLen)
	stats := ResourceStats{
		Resource: ResourceMetadata{
			OrgID:          ref.OrgID,
			ProjectID:      ref.ProjectID,
			ResourceType:   ref.ResourceType,
			ResourceID:     ref.ResourceID,
			SourceMIMEType: strings.TrimSpace(sourceMIMEType),
			IndexedAt:      now.UTC().Format(time.RFC3339),
			ByteLength:     frontmatterLen + len(body),
		},
		Index: ResourceIndex{Sections: sections},
	}
	rendered, err := renderIndexedFrontmatter(frontmatter, stats)
	if err != nil {
		return IndexedResource{}, err
	}
	return IndexedResource{Ref: ref, Markdown: rendered + body, Stats: stats}, nil
}

func ParseStats(markdownPrefix string) (ResourceStats, error) {
	frontmatter, _, err := splitFrontmatter(markdownPrefix)
	if err != nil {
		return ResourceStats{}, err
	}
	rawResource, ok := frontmatter["storage"]
	if !ok {
		return ResourceStats{}, invalid("storage.stats.missing", "storage frontmatter is missing")
	}
	rawIndex, ok := frontmatter["index"]
	if !ok {
		return ResourceStats{}, invalid("storage.stats.missing", "index frontmatter is missing")
	}
	resourceBytes, err := yaml.Marshal(rawResource)
	if err != nil {
		return ResourceStats{}, dependency("storage.stats.marshal", "Storage stats parse failed", err)
	}
	indexBytes, err := yaml.Marshal(rawIndex)
	if err != nil {
		return ResourceStats{}, dependency("storage.stats.marshal", "Storage stats parse failed", err)
	}
	var stats ResourceStats
	if err := yaml.Unmarshal(resourceBytes, &stats.Resource); err != nil {
		return ResourceStats{}, invalid("storage.stats.invalid", "storage metadata is invalid")
	}
	if err := yaml.Unmarshal(indexBytes, &stats.Index); err != nil {
		return ResourceStats{}, invalid("storage.stats.invalid", "storage index is invalid")
	}
	return stats, nil
}

func splitFrontmatter(markdown string) (map[string]any, string, error) {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	if !strings.HasPrefix(markdown, "---\n") {
		return map[string]any{}, markdown, nil
	}
	rest := strings.TrimPrefix(markdown, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return nil, "", invalid("storage.frontmatter.unclosed", "markdown frontmatter is unclosed")
	}
	rawFrontmatter := rest[:idx]
	body := rest[idx+len("\n---\n"):]
	out := map[string]any{}
	if strings.TrimSpace(rawFrontmatter) == "" {
		return out, body, nil
	}
	if err := yaml.Unmarshal([]byte(rawFrontmatter), &out); err != nil {
		return nil, "", invalid("storage.frontmatter.invalid", "markdown frontmatter is invalid")
	}
	return out, body, nil
}

func renderIndexedFrontmatter(frontmatter map[string]any, stats ResourceStats) (string, error) {
	out := map[string]any{}
	for key, value := range frontmatter {
		out[key] = value
	}
	out["storage"] = stats.Resource
	out["index"] = stats.Index
	raw, err := yaml.Marshal(out)
	if err != nil {
		return "", dependency("storage.frontmatter.render", "Storage frontmatter render failed", err)
	}
	return "---\n" + string(raw) + "---\n\n", nil
}

func scanSections(markdown string, offset int) []SectionIndex {
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var sections []SectionIndex
	position := offset
	for scanner.Scan() {
		line := scanner.Text()
		lineLen := len(line) + 1
		if level, title, ok := markdownHeader(line); ok {
			if len(sections) > 0 {
				sections[len(sections)-1].EndByte = position - 1
			}
			sections = append(sections, SectionIndex{
				Title:     title,
				Level:     level,
				Anchor:    anchorFor(title),
				StartByte: position,
			})
		}
		position += lineLen
	}
	if len(sections) > 0 {
		sections[len(sections)-1].EndByte = offset + len(markdown) - 1
	}
	return sections
}

func markdownHeader(line string) (int, string, bool) {
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	title := strings.TrimSpace(line[level+1:])
	if title == "" {
		return 0, "", false
	}
	return level, title, true
}

func anchorFor(title string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	anchor := strings.Trim(b.String(), "-")
	if anchor != "" {
		return anchor
	}
	sum := sha1.Sum([]byte(title))
	return "section-" + hex.EncodeToString(sum[:4])
}

func normalizeMarkdownBody(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.TrimLeft(body, "\n")
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return strings.TrimRight(body, "\n") + "\n"
}

func headerAt(markdown string, offset int, title string) bool {
	if offset < 0 || offset >= len(markdown) {
		return false
	}
	lineEnd := bytes.IndexByte([]byte(markdown[offset:]), '\n')
	if lineEnd < 0 {
		lineEnd = len(markdown) - offset
	}
	line := markdown[offset : offset+lineEnd]
	return strings.Contains(line, title)
}

func byteRange(start, end int) string {
	return "bytes=" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
}

func validateIndexOffsets(markdown string, sections []SectionIndex) error {
	for _, section := range sections {
		if !headerAt(markdown, section.StartByte, section.Title) {
			return fmt.Errorf("section %q start byte %d does not point at its heading", section.Title, section.StartByte)
		}
		if section.EndByte < section.StartByte || section.EndByte >= len(markdown) {
			return fmt.Errorf("section %q has invalid byte range %s", section.Title, byteRange(section.StartByte, section.EndByte))
		}
	}
	return nil
}
