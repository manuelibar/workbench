package storage

import (
	"path"
	"strings"
)

type ResourceRef struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

func NewResourceRef(orgID, projectID, resourceType, resourceID string) (ResourceRef, error) {
	ref := ResourceRef{
		OrgID:        strings.TrimSpace(orgID),
		ProjectID:    strings.TrimSpace(projectID),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
	}
	if err := ref.Validate(); err != nil {
		return ResourceRef{}, err
	}
	return ref, nil
}

func (r ResourceRef) Validate() error {
	for name, value := range map[string]string{
		"org_id":        r.OrgID,
		"project_id":    r.ProjectID,
		"resource_type": r.ResourceType,
		"resource_id":   r.ResourceID,
	} {
		if strings.TrimSpace(value) == "" {
			return invalid("storage.ref.required", "%s is required", name)
		}
		if !safePathPart(value) {
			return invalid("storage.ref.invalid", "%s is invalid", name)
		}
	}
	return nil
}

func (r ResourceRef) Key() string {
	return path.Join(r.OrgID, r.ProjectID, r.ResourceType, r.ResourceID+".md")
}

func (r ResourceRef) URI() string {
	return "storage:///" + r.Key()
}

func InboxKey(orgID, projectID, filename string) (string, error) {
	orgID = strings.TrimSpace(orgID)
	projectID = strings.TrimSpace(projectID)
	filename = strings.TrimSpace(filename)
	if orgID == "" || projectID == "" || filename == "" {
		return "", invalid("storage.inbox.required", "org_id, project_id, and filename are required")
	}
	if !safePathPart(orgID) || !safePathPart(projectID) || !safeFilename(filename) {
		return "", invalid("storage.inbox.invalid", "inbox key contains an invalid path segment")
	}
	return path.Join("inbox", orgID, projectID, filename), nil
}

func Prefix(orgID, projectID, resourceType string) (string, error) {
	orgID = strings.TrimSpace(orgID)
	projectID = strings.TrimSpace(projectID)
	resourceType = strings.TrimSpace(resourceType)
	if orgID == "" || projectID == "" || resourceType == "" {
		return "", invalid("storage.prefix.required", "org_id, project_id, and resource_type are required")
	}
	for name, value := range map[string]string{
		"org_id":        orgID,
		"project_id":    projectID,
		"resource_type": resourceType,
	} {
		if !safePathPart(value) {
			return "", invalid("storage.prefix.invalid", "%s is invalid", name)
		}
	}
	return path.Join(orgID, projectID, resourceType) + "/", nil
}

func RefFromKey(key string) (ResourceRef, bool) {
	key = strings.Trim(strings.TrimSpace(key), "/")
	parts := strings.Split(key, "/")
	if len(parts) != 4 || !strings.HasSuffix(parts[3], ".md") {
		return ResourceRef{}, false
	}
	id := strings.TrimSuffix(parts[3], ".md")
	ref := ResourceRef{
		OrgID:        parts[0],
		ProjectID:    parts[1],
		ResourceType: parts[2],
		ResourceID:   id,
	}
	return ref, ref.Validate() == nil
}

func safeFilename(value string) bool {
	if value == "." || value == ".." || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return false
	}
	return safePathPart(value)
}

func safePathPart(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return !strings.Contains(value, "..")
}
