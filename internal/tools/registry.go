package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
)

// RegisterAllTools resolves the given profiles and registers the corresponding
// tools with the MCP server. Returns an error if profile resolution fails.
func RegisterAllTools(s *mcp.Server, client *api.Client, profiles []string) error {
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
		registerReadwiseToolsFiltered(s, client, activeTools)
	}
	if profileSet["reader"] {
		registerReaderToolsFiltered(s, client, activeTools)
	}
	if profileSet["write"] {
		registerWriteToolsFiltered(s, client, activeTools)
	}
	if profileSet["video"] {
		registerVideoToolsFiltered(s, client, activeTools)
	}
	if profileSet["destructive"] {
		registerDestructiveToolsFiltered(s, client, activeTools)
	}

	return nil
}

// registerReadwiseToolsFiltered registers readwise tools if they're in the active set.
func registerReadwiseToolsFiltered(s *mcp.Server, client *api.Client, activeTools map[string]bool) {
	RegisterReadwiseTools(s, client)
	if activeTools["search_highlights"] {
		RegisterSearchHighlightsTool(s, client)
	}
}

// registerReaderToolsFiltered registers reader tools if they're in the active set.
func registerReaderToolsFiltered(s *mcp.Server, client *api.Client, activeTools map[string]bool) {
	RegisterReaderTools(s, client)
	if activeTools["search_documents"] {
		RegisterSearchDocumentsTool(s, client)
	}
}

// registerWriteToolsFiltered is a placeholder until write tools are implemented.
func registerWriteToolsFiltered(s *mcp.Server, client *api.Client, activeTools map[string]bool) {
	// Write tools will be registered in Phase 7
}

// registerVideoToolsFiltered is a placeholder until video tools are implemented.
func registerVideoToolsFiltered(s *mcp.Server, client *api.Client, activeTools map[string]bool) {
	// Video tools will be registered in Phase 8
}

// registerDestructiveToolsFiltered is a placeholder until destructive tools are implemented.
func registerDestructiveToolsFiltered(s *mcp.Server, client *api.Client, activeTools map[string]bool) {
	// Destructive tools will be registered in Phase 9
}
