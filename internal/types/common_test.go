package types

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear any env vars that might interfere
	for _, key := range []string{"READWISE_PROFILES", "PORT", "LOG_LEVEL", "CACHE_MAX_SIZE_MB", "CACHE_TTL_SECONDS", "CACHE_ENABLED"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	cfg := LoadConfig()

	if len(cfg.Profiles) != 1 || cfg.Profiles[0] != "readwise" {
		t.Errorf("Profiles = %v, want [readwise]", cfg.Profiles)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.CacheMaxSizeMB != 128 {
		t.Errorf("CacheMaxSizeMB = %d, want 128", cfg.CacheMaxSizeMB)
	}
	if cfg.CacheTTLSeconds != 300 {
		t.Errorf("CacheTTLSeconds = %d, want 300", cfg.CacheTTLSeconds)
	}
	if !cfg.CacheEnabled {
		t.Error("CacheEnabled = false, want true")
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv("READWISE_PROFILES", "reader,write")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("CACHE_MAX_SIZE_MB", "256")
	t.Setenv("CACHE_TTL_SECONDS", "600")
	t.Setenv("CACHE_ENABLED", "false")

	cfg := LoadConfig()

	if len(cfg.Profiles) != 2 || cfg.Profiles[0] != "reader" || cfg.Profiles[1] != "write" {
		t.Errorf("Profiles = %v, want [reader write]", cfg.Profiles)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.CacheMaxSizeMB != 256 {
		t.Errorf("CacheMaxSizeMB = %d, want 256", cfg.CacheMaxSizeMB)
	}
	if cfg.CacheTTLSeconds != 600 {
		t.Errorf("CacheTTLSeconds = %d, want 600", cfg.CacheTTLSeconds)
	}
	if cfg.CacheEnabled {
		t.Error("CacheEnabled = true, want false")
	}
}

func TestLoadConfigProfileParsing(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected []string
	}{
		{"single", "readwise", []string{"readwise"}},
		{"multiple", "readwise,reader,write", []string{"readwise", "reader", "write"}},
		{"with spaces", " readwise , reader ", []string{"readwise", "reader"}},
		{"empty entries", "readwise,,reader", []string{"readwise", "reader"}},
		{"all", "all", []string{"all"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("READWISE_PROFILES", tt.env)
			cfg := LoadConfig()
			if len(cfg.Profiles) != len(tt.expected) {
				t.Fatalf("len(Profiles) = %d, want %d", len(cfg.Profiles), len(tt.expected))
			}
			for i, p := range cfg.Profiles {
				if p != tt.expected[i] {
					t.Errorf("Profiles[%d] = %q, want %q", i, p, tt.expected[i])
				}
			}
		})
	}
}

func TestConfigAddr(t *testing.T) {
	cfg := Config{Port: 8080}
	if addr := cfg.Addr(); addr != ":8080" {
		t.Errorf("Addr() = %q, want %q", addr, ":8080")
	}
}
