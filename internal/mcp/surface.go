package mcp

import (
	"context"
	"fmt"
	"sync"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/manuelibar/workbench/internal/artifacts"
	mcpresources "github.com/manuelibar/workbench/internal/mcp/resources"
	"github.com/manuelibar/workbench/internal/mcp/tools"
	"github.com/manuelibar/workbench/internal/set"
)

type surfaceSynchronizer struct {
	mu        sync.Mutex
	tools     set.Set[string]
	resources set.Set[string]
}

func newSurfaceSynchronizer() *surfaceSynchronizer {
	return &surfaceSynchronizer{
		tools:     set.New[string](),
		resources: set.New[string](),
	}
}

func (sf *surfaceSynchronizer) ChangedCategories(surface tools.CapabilitySurface) []string {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	desiredTools := set.New[string]()
	for _, tool := range surface.Tools {
		desiredTools.Add(tool.Name)
	}
	desiredResources := set.New[string]()
	for _, resource := range surface.Resources {
		desiredResources.Add(resource.URI)
	}

	var changed []string
	if !sf.tools.Equal(desiredTools) {
		changed = append(changed, "tools")
	}
	if !sf.resources.Equal(desiredResources) {
		changed = append(changed, "resources")
	}
	return changed
}

func (sf *surfaceSynchronizer) Synchronize(server *Server, surface tools.CapabilitySurface) []string {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	desiredTools := set.New[string]()
	for _, tool := range surface.Tools {
		desiredTools.Add(tool.Name)
	}
	desiredResources := set.New[string]()
	for _, resource := range surface.Resources {
		desiredResources.Add(resource.URI)
	}

	var changed []string
	if !sf.tools.Equal(desiredTools) {
		changed = append(changed, "tools")
	}
	if !sf.resources.Equal(desiredResources) {
		changed = append(changed, "resources")
	}

	for name := range sf.tools {
		if desiredTools.Has(name) {
			continue
		}
		server.sdk.RemoveTools(name)
		sf.tools.Delete(name)
	}
	for uri := range sf.resources {
		if desiredResources.Has(uri) {
			continue
		}
		server.sdk.RemoveResources(uri)
		sf.resources.Delete(uri)
	}
	for _, summary := range surface.Tools {
		if sf.tools.Has(summary.Name) {
			continue
		}
		def, ok := tools.DefaultRegistry().ByName(summary.Name)
		if !ok {
			panic(fmt.Sprintf("unknown tool %q", summary.Name))
		}
		tools.Bind(def, server.sdk, server)
		sf.tools.Add(summary.Name)
	}
	for _, summary := range surface.Resources {
		if sf.resources.Has(summary.URI) {
			continue
		}
		def, ok := mcpresources.DefaultRegistry().ByURI(summary.URI)
		if !ok {
			panic(fmt.Sprintf("unknown resource %q", summary.URI))
		}
		switch {
		case summary.URI == mcpresources.ScopeURI:
			server.sdk.AddResource(&mcpsdk.Resource{
				URI:         summary.URI,
				Name:        def.Name(),
				Title:       def.Title(),
				Description: def.Description(),
				MIMEType:    def.MIMEType(),
			}, server.readScopeResource)
		case artifactIDFromURI(summary.URI) != "":
			id := artifactIDFromURI(summary.URI)
			if artifact, err := server.artifacts.GetContext(context.Background(), id); err == nil {
				def = mcpresources.NewArtifactResource(mcpresources.Artifact{
					ID:     artifact.ID,
					Type:   artifact.Type,
					Title:  artifact.Title,
					Status: artifact.Status,
				})
			}
			server.sdk.AddResource(&mcpsdk.Resource{
				URI:         summary.URI,
				Name:        def.Name(),
				Title:       def.Title(),
				Description: def.Description(),
				MIMEType:    def.MIMEType(),
			}, server.readArtifactResource)
		default:
			panic(fmt.Sprintf("unknown resource %q", summary.URI))
		}
		sf.resources.Add(summary.URI)
	}
	return changed
}

func (sf *surfaceSynchronizer) RefreshArtifactResource(server *Server, artifact artifacts.Summary) {
	uri := mcpresources.ArtifactURI(artifact.ID)
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if !sf.resources.Has(uri) {
		return
	}
	def := mcpresources.NewArtifactResource(mcpresources.Artifact{
		ID:     artifact.ID,
		Type:   artifact.Type,
		Title:  artifact.Title,
		Status: artifact.Status,
	})
	server.sdk.AddResource(&mcpsdk.Resource{
		URI:         def.URI(),
		Name:        def.Name(),
		Title:       def.Title(),
		Description: def.Description(),
		MIMEType:    def.MIMEType(),
	}, server.readArtifactResource)
}
