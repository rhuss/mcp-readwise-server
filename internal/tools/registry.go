package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
)

// RegisterAllTools resolves the given profiles and registers the corresponding
// tools with the MCP server. Returns an error if profile resolution fails.
func RegisterAllTools(s *mcp.Server, client *api.Client, cm *cache.Manager, profiles []string) error {
	resolved, err := ResolveProfiles(profiles)
	if err != nil {
		return err
	}

	activeTools := make(map[string]bool)
	for _, tool := range ToolsForProfiles(resolved) {
		activeTools[tool] = true
	}

	profileSet := make(map[string]bool)
	for _, p := range resolved {
		profileSet[p] = true
	}

	// Register tools based on active profiles
	if profileSet["readwise"] {
		RegisterReadwiseTools(s, client)
		if activeTools["search_highlights"] {
			RegisterSearchHighlightsTool(s, client)
		}
	}
	if profileSet["reader"] {
		RegisterReaderTools(s, client)
		if activeTools["search_documents"] {
			RegisterSearchDocumentsTool(s, client)
		}
	}
	if profileSet["write"] {
		RegisterWriteTools(s, client, cm)
	}
	if profileSet["video"] {
		RegisterVideoTools(s, client, cm)
	}
	if profileSet["destructive"] {
		RegisterDestructiveTools(s, client, cm)
	}

	return nil
}
