package tools

import (
	"fmt"
	"slices"
	"strings"
)

// ProfileType indicates whether a profile provides read access or modifies data.
type ProfileType string

const (
	ProfileTypeRead     ProfileType = "read"
	ProfileTypeModifier ProfileType = "modifier"
)

// Profile defines a named group of MCP tools with dependencies.
type Profile struct {
	Name         string
	Type         ProfileType
	Dependencies []string // Required profiles that must also be active
	ToolNames    []string // Names of tools belonging to this profile
}

// baseProfiles defines the 5 base profiles with their tool assignments.
var baseProfiles = map[string]Profile{
	"readwise": {
		Name: "readwise",
		Type: ProfileTypeRead,
		ToolNames: []string{
			"list_sources", "get_source",
			"list_highlights", "get_highlight",
			"export_highlights", "get_daily_review",
			"list_source_tags", "list_highlight_tags",
			"search_highlights",
		},
	},
	"reader": {
		Name: "reader",
		Type: ProfileTypeRead,
		ToolNames: []string{
			"list_documents", "get_document",
			"list_reader_tags",
			"search_documents",
		},
	},
	"write": {
		Name:         "write",
		Type:         ProfileTypeModifier,
		Dependencies: []string{"readwise|reader"}, // requires at least one read profile
		ToolNames: []string{
			"save_document", "update_document",
			"create_highlight", "update_highlight",
			"add_source_tag", "add_highlight_tag",
			"bulk_create_highlights",
		},
	},
	"video": {
		Name:         "video",
		Type:         ProfileTypeModifier,
		Dependencies: []string{"reader"},
		ToolNames: []string{
			"list_videos", "get_video", "get_video_position",
			"update_video_position", "create_video_highlight",
		},
	},
	"destructive": {
		Name:         "destructive",
		Type:         ProfileTypeModifier,
		Dependencies: []string{"readwise|reader"}, // requires at least one read profile
		ToolNames: []string{
			"delete_highlight", "delete_highlight_tag",
			"delete_source_tag", "delete_document",
		},
	},
}

// shortcuts maps shortcut names to profile lists.
var shortcuts = map[string][]string{
	"basic": {"reader", "write"},
	"all":   {"readwise", "reader", "write", "video", "destructive"},
}

// ResolveProfiles takes a list of profile names (possibly including shortcuts),
// expands shortcuts, deduplicates, validates dependencies, and returns the
// resolved set of active profile names.
func ResolveProfiles(names []string) ([]string, error) {
	// Expand shortcuts
	expanded := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if shortcut, ok := shortcuts[name]; ok {
			expanded = append(expanded, shortcut...)
		} else {
			expanded = append(expanded, name)
		}
	}

	// Deduplicate while preserving order
	seen := make(map[string]bool)
	resolved := make([]string, 0, len(expanded))
	for _, name := range expanded {
		if !seen[name] {
			seen[name] = true
			resolved = append(resolved, name)
		}
	}

	// Validate all names are known profiles
	for _, name := range resolved {
		if _, ok := baseProfiles[name]; !ok {
			return nil, fmt.Errorf("unknown profile: %q", name)
		}
	}

	// Validate dependencies
	for _, name := range resolved {
		profile := baseProfiles[name]
		for _, dep := range profile.Dependencies {
			if !dependencySatisfied(dep, seen) {
				return nil, fmt.Errorf("profile %q requires %s, but it is not enabled", name, formatDependency(dep))
			}
		}
	}

	return resolved, nil
}

// dependencySatisfied checks if a dependency string is satisfied.
// Dependencies can use "|" to indicate "at least one of".
func dependencySatisfied(dep string, activeSet map[string]bool) bool {
	alternatives := strings.Split(dep, "|")
	for _, alt := range alternatives {
		if activeSet[strings.TrimSpace(alt)] {
			return true
		}
	}
	return false
}

// formatDependency returns a human-readable dependency description.
func formatDependency(dep string) string {
	alternatives := strings.Split(dep, "|")
	if len(alternatives) == 1 {
		return fmt.Sprintf("profile %q", alternatives[0])
	}
	quoted := make([]string, len(alternatives))
	for i, a := range alternatives {
		quoted[i] = fmt.Sprintf("%q", strings.TrimSpace(a))
	}
	return fmt.Sprintf("one of profiles %s", strings.Join(quoted, " or "))
}

// ToolsForProfiles returns the deduplicated list of tool names that should be
// registered for the given active profiles.
func ToolsForProfiles(profiles []string) []string {
	seen := make(map[string]bool)
	var tools []string

	for _, name := range profiles {
		profile, ok := baseProfiles[name]
		if !ok {
			continue
		}
		for _, tool := range profile.ToolNames {
			if !seen[tool] {
				seen[tool] = true
				tools = append(tools, tool)
			}
		}
	}

	return tools
}

// ProfileForTool returns the profile that owns a given tool name.
func ProfileForTool(toolName string) (string, bool) {
	for name, profile := range baseProfiles {
		if slices.Contains(profile.ToolNames, toolName) {
			return name, true
		}
	}
	return "", false
}
