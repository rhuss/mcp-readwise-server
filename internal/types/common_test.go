package types

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear any env vars that might interfere
	for _, key := range []string{"READWISE_PROFILES", "PORT", "LOG_LEVEL", "CACHE_MAX_SIZE_MB", "CACHE_TTL_SECONDS", "CACHE_ENABLED", "TLS_CERT_FILE", "TLS_KEY_FILE", "TLS_PORT"} {
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

func TestTLSDefaults(t *testing.T) {
	for _, key := range []string{"TLS_CERT_FILE", "TLS_KEY_FILE", "TLS_PORT"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	cfg := LoadConfig()

	if cfg.TLSCertFile != "" {
		t.Errorf("TLSCertFile = %q, want empty", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "" {
		t.Errorf("TLSKeyFile = %q, want empty", cfg.TLSKeyFile)
	}
	if cfg.TLSPort != 8443 {
		t.Errorf("TLSPort = %d, want 8443", cfg.TLSPort)
	}
	if cfg.TLSEnabled() {
		t.Error("TLSEnabled() = true, want false")
	}
}

func TestTLSEnvOverrides(t *testing.T) {
	t.Setenv("TLS_CERT_FILE", "/etc/tls/tls.crt")
	t.Setenv("TLS_KEY_FILE", "/etc/tls/tls.key")
	t.Setenv("TLS_PORT", "9443")

	cfg := LoadConfig()

	if cfg.TLSCertFile != "/etc/tls/tls.crt" {
		t.Errorf("TLSCertFile = %q, want %q", cfg.TLSCertFile, "/etc/tls/tls.crt")
	}
	if cfg.TLSKeyFile != "/etc/tls/tls.key" {
		t.Errorf("TLSKeyFile = %q, want %q", cfg.TLSKeyFile, "/etc/tls/tls.key")
	}
	if cfg.TLSPort != 9443 {
		t.Errorf("TLSPort = %d, want 9443", cfg.TLSPort)
	}
}

func TestTLSEnabled(t *testing.T) {
	tests := []struct {
		name     string
		cert     string
		key      string
		expected bool
	}{
		{"both set", "/cert.pem", "/key.pem", true},
		{"neither set", "", "", false},
		{"cert only", "/cert.pem", "", false},
		{"key only", "", "/key.pem", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{TLSCertFile: tt.cert, TLSKeyFile: tt.key}
			if got := cfg.TLSEnabled(); got != tt.expected {
				t.Errorf("TLSEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTLSAddr(t *testing.T) {
	cfg := Config{TLSPort: 8443}
	if addr := cfg.TLSAddr(); addr != ":8443" {
		t.Errorf("TLSAddr() = %q, want %q", addr, ":8443")
	}
}

func TestValidateTLS(t *testing.T) {
	// Create temp cert and key files for file existence tests
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile.Name())
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())
	keyFile.Close()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "TLS disabled (no config)",
			cfg:     Config{Port: 8080, TLSPort: 8443},
			wantErr: false,
		},
		{
			name:    "cert without key",
			cfg:     Config{Port: 8080, TLSCertFile: certFile.Name(), TLSPort: 8443},
			wantErr: true,
			errMsg:  "TLS configuration incomplete",
		},
		{
			name:    "key without cert",
			cfg:     Config{Port: 8080, TLSKeyFile: keyFile.Name(), TLSPort: 8443},
			wantErr: true,
			errMsg:  "TLS configuration incomplete",
		},
		{
			name:    "cert file missing",
			cfg:     Config{Port: 8080, TLSCertFile: "/nonexistent/cert.pem", TLSKeyFile: keyFile.Name(), TLSPort: 8443},
			wantErr: true,
			errMsg:  "TLS certificate file not found",
		},
		{
			name:    "key file missing",
			cfg:     Config{Port: 8080, TLSCertFile: certFile.Name(), TLSKeyFile: "/nonexistent/key.pem", TLSPort: 8443},
			wantErr: true,
			errMsg:  "TLS key file not found",
		},
		{
			name:    "port conflict",
			cfg:     Config{Port: 8080, TLSCertFile: certFile.Name(), TLSKeyFile: keyFile.Name(), TLSPort: 8080},
			wantErr: true,
			errMsg:  "conflicts with HTTP port",
		},
		{
			name:    "valid TLS config",
			cfg:     Config{Port: 8080, TLSCertFile: certFile.Name(), TLSKeyFile: keyFile.Name(), TLSPort: 8443},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateTLS()
			if tt.wantErr {
				if err == nil {
					t.Error("ValidateTLS() = nil, want error")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateTLS() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidateTLS() = %v, want nil", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
