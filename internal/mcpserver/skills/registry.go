package skills

// ProjectContext carries the selected project's data into dynamic skill rendering.
// All fields are empty when no project is selected.
type ProjectContext struct {
	Name         string
	Description  string
	SystemPrompt string
}

// File is a single file within a Bundle.
type File struct {
	RelPath  string
	MIMEType string
	// Content returns the file bytes rendered for the given project context.
	// Static files ignore ctx entirely.
	Content func(ctx ProjectContext) []byte
}

// Bundle is a named, versioned skill bundle.
type Bundle struct {
	Name        string
	Description string
	Version     string
	Files       []File
}

// SkillRegistry is the storage contract for skill bundles.
type SkillRegistry interface {
	// All returns every registered bundle.
	All() []Bundle
	// Get returns the bundle with the given name, or (zero, false).
	Get(name string) (Bundle, bool)
	// For returns bundles visible given whether a project is currently selected.
	For(hasProject bool) []Bundle
}
