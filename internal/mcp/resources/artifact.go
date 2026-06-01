package resources

import "strings"

type SelectedArtifact struct {
	ID     string
	Type   string
	Title  string
	Status string
}

type SelectedArtifactResource struct {
	artifact SelectedArtifact
}

func NewSelectedArtifactResource(artifact SelectedArtifact) *SelectedArtifactResource {
	return &SelectedArtifactResource{artifact: SelectedArtifact{
		ID:     strings.TrimSpace(artifact.ID),
		Type:   strings.TrimSpace(artifact.Type),
		Title:  strings.TrimSpace(artifact.Title),
		Status: strings.TrimSpace(artifact.Status),
	}}
}

func (r *SelectedArtifactResource) URI() string {
	if r.artifact.ID == "" {
		return ""
	}
	return ArtifactURI(r.artifact.ID)
}

func (r *SelectedArtifactResource) Name() string {
	if r.artifact.Title != "" {
		return r.artifact.Title
	}
	if r.artifact.ID != "" {
		return r.artifact.ID
	}
	return "artifact"
}

func (r *SelectedArtifactResource) Title() string {
	if r.artifact.Title != "" {
		return r.artifact.Title
	}
	return "Artifact"
}

func (r *SelectedArtifactResource) Description() string {
	var parts []string
	if r.artifact.Type != "" {
		parts = append(parts, r.artifact.Type)
	}
	if r.artifact.Status != "" {
		parts = append(parts, r.artifact.Status)
	}
	if len(parts) == 0 {
		return "Read the selected artifact Markdown resource."
	}
	return "Read the selected " + strings.Join(parts, " ") + " artifact Markdown resource."
}

func (r *SelectedArtifactResource) MIMEType() string {
	return "text/markdown"
}

func (r *SelectedArtifactResource) Group() string {
	return "artifacts"
}

func (r *SelectedArtifactResource) Visibility() Visibility {
	return VisibleArtifactSelected
}

type ArtifactTemplate struct{}

func NewArtifactTemplate() *ArtifactTemplate {
	return &ArtifactTemplate{}
}

func (t *ArtifactTemplate) URITemplate() string {
	return ArtifactTemplateURI
}

func (t *ArtifactTemplate) Name() string {
	return "artifact"
}

func (t *ArtifactTemplate) Title() string {
	return "Artifact"
}

func (t *ArtifactTemplate) Description() string {
	return "Read an artifact Markdown resource by stable id."
}

func (t *ArtifactTemplate) MIMEType() string {
	return "text/markdown"
}

func (t *ArtifactTemplate) Group() string {
	return "artifacts"
}

func (t *ArtifactTemplate) Visibility() Visibility {
	return VisibleAlways
}
