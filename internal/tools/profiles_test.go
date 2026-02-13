package tools

import (
	"testing"
)

func TestBaseProfileDefinitions(t *testing.T) {
	tests := []struct {
		profile   string
		toolCount int
	}{
		{"readwise", 9},
		{"reader", 4},
		{"write", 7},
		{"video", 5},
		{"destructive", 4},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			p, ok := baseProfiles[tt.profile]
			if !ok {
				t.Fatalf("profile %q not defined", tt.profile)
			}
			if len(p.ToolNames) != tt.toolCount {
				t.Errorf("profile %q has %d tools, want %d", tt.profile, len(p.ToolNames), tt.toolCount)
			}
		})
	}

	// Total should be 29
	total := 0
	for _, p := range baseProfiles {
		total += len(p.ToolNames)
	}
	if total != 29 {
		t.Errorf("total tools = %d, want 29", total)
	}
}

func TestShortcutExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "basic expands to reader+write",
			input:    []string{"basic"},
			expected: []string{"reader", "write"},
		},
		{
			name:     "all expands to all 5 profiles",
			input:    []string{"all"},
			expected: []string{"readwise", "reader", "write", "video", "destructive"},
		},
		{
			name:     "non-shortcut passes through",
			input:    []string{"readwise"},
			expected: []string{"readwise"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveProfiles(tt.input)
			if err != nil {
				t.Fatalf("ResolveProfiles error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("len(result) = %d, want %d: %v", len(result), len(tt.expected), result)
			}
			for i, p := range result {
				if p != tt.expected[i] {
					t.Errorf("result[%d] = %q, want %q", i, p, tt.expected[i])
				}
			}
		})
	}
}

func TestDependencyValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr bool
	}{
		{
			name:    "write without read fails",
			input:   []string{"write"},
			wantErr: true,
		},
		{
			name:    "write with readwise succeeds",
			input:   []string{"readwise", "write"},
			wantErr: false,
		},
		{
			name:    "write with reader succeeds",
			input:   []string{"reader", "write"},
			wantErr: false,
		},
		{
			name:    "video without reader fails",
			input:   []string{"video"},
			wantErr: true,
		},
		{
			name:    "video with reader succeeds",
			input:   []string{"reader", "video"},
			wantErr: false,
		},
		{
			name:    "destructive without read fails",
			input:   []string{"destructive"},
			wantErr: true,
		},
		{
			name:    "destructive with readwise succeeds",
			input:   []string{"readwise", "destructive"},
			wantErr: false,
		},
		{
			name:    "basic self-resolves (reader+write)",
			input:   []string{"basic"},
			wantErr: false,
		},
		{
			name:    "all self-resolves",
			input:   []string{"all"},
			wantErr: false,
		},
		{
			name:    "unknown profile fails",
			input:   []string{"unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveProfiles(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProfiles(%v) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestToolFiltering(t *testing.T) {
	tests := []struct {
		name      string
		profiles  []string
		wantCount int
	}{
		{"readwise only", []string{"readwise"}, 9},
		{"reader only", []string{"reader"}, 4},
		{"readwise+reader", []string{"readwise", "reader"}, 13},
		{"all profiles", []string{"readwise", "reader", "write", "video", "destructive"}, 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := ToolsForProfiles(tt.profiles)
			if len(tools) != tt.wantCount {
				t.Errorf("ToolsForProfiles(%v) returned %d tools, want %d", tt.profiles, len(tools), tt.wantCount)
			}

			// Verify no duplicates
			seen := make(map[string]bool)
			for _, tool := range tools {
				if seen[tool] {
					t.Errorf("duplicate tool: %q", tool)
				}
				seen[tool] = true
			}
		})
	}
}

func TestToolFilteringDeduplication(t *testing.T) {
	// If profiles somehow share tools, they should be deduplicated
	tools := ToolsForProfiles([]string{"readwise", "readwise"})
	if len(tools) != 9 {
		t.Errorf("expected 9 tools, got %d (duplicate profile should not duplicate tools)", len(tools))
	}
}

func TestProfileForTool(t *testing.T) {
	tests := []struct {
		tool     string
		expected string
		found    bool
	}{
		{"list_sources", "readwise", true},
		{"list_documents", "reader", true},
		{"save_document", "write", true},
		{"list_videos", "video", true},
		{"delete_highlight", "destructive", true},
		{"nonexistent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			profile, found := ProfileForTool(tt.tool)
			if found != tt.found {
				t.Errorf("ProfileForTool(%q) found = %v, want %v", tt.tool, found, tt.found)
			}
			if profile != tt.expected {
				t.Errorf("ProfileForTool(%q) = %q, want %q", tt.tool, profile, tt.expected)
			}
		})
	}
}
