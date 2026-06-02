package resources

import "strings"

type Artifact struct {
	ID     string
	Type   string
	Title  string
	Status string
}

type ArtifactResource struct {
	artifact Artifact
}

func init() {
	register(NewArtifactResource(Artifact{}))
}

func NewArtifactResource(artifact Artifact) *ArtifactResource {
	return &ArtifactResource{artifact: Artifact{
		ID:     strings.TrimSpace(artifact.ID),
		Type:   strings.TrimSpace(artifact.Type),
		Title:  strings.TrimSpace(artifact.Title),
		Status: strings.TrimSpace(artifact.Status),
	}}
}

func (r *ArtifactResource) URI() string {
	if r.artifact.ID == "" {
		return ""
	}
	return ArtifactURI(r.artifact.ID)
}

func (r *ArtifactResource) Name() string {
	if r.artifact.Title != "" {
		return r.artifact.Title
	}
	if r.artifact.ID != "" {
		return r.artifact.ID
	}
	return "artifact"
}

func (r *ArtifactResource) Title() string {
	if r.artifact.Title != "" {
		return r.artifact.Title
	}
	return "Artifact"
}

func (r *ArtifactResource) Description() string {
	var parts []string
	if r.artifact.Type != "" {
		parts = append(parts, r.artifact.Type)
	}
	if r.artifact.Status != "" {
		parts = append(parts, r.artifact.Status)
	}
	if len(parts) == 0 {
		return "Read artifact Markdown."
	}
	return "Read the " + strings.Join(parts, " ") + " artifact Markdown."
}

func (r *ArtifactResource) MIMEType() string {
	return "text/markdown"
}

func (r *ArtifactResource) Group() string {
	return "artifacts"
}

func (r *ArtifactResource) Visibility() Visibility {
	return VisibleArtifactScoped
}
