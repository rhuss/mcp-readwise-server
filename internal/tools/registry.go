package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
)

// RegisterAllTools registers all tools with the MCP server based on the active profiles.
// In the initial implementation, this registers all readwise tools unconditionally.
// Profile filtering will be added in Phase 5 (US4).
func RegisterAllTools(s *mcp.Server, client *api.Client, profiles []string) {
	profileSet := make(map[string]bool)
	for _, p := range profiles {
		profileSet[p] = true
	}

	if profileSet["readwise"] || profileSet["all"] {
		RegisterReadwiseTools(s, client)
		RegisterSearchHighlightsTool(s, client)
	}

	if profileSet["reader"] || profileSet["all"] || profileSet["basic"] {
		RegisterReaderTools(s, client)
		RegisterSearchDocumentsTool(s, client)
	}
}
