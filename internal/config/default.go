package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ipinfo", "config.yaml")
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() *Config {
	return &Config{
		Cache: CacheConfig{
			TTL:     5 * time.Minute,
			MaxSize: 1000,
		},
		CircuitBreaker: CircuitBreakerConfig{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
		Providers: []ProviderConfig{
			// Tier 1
			{Name: "ipinfo.io", BaseURL: "https://ipinfo.io", Priority: 1, Token: "", Enabled: true, Timeout: 5 * time.Second},
			// Tier 2
			{Name: "ip-api.com", BaseURL: "http://ip-api.com/json", Priority: 2, Enabled: true, Timeout: 5 * time.Second},
			{Name: "ipapi.co", BaseURL: "https://ipapi.co", Priority: 2, Enabled: true, Timeout: 5 * time.Second},
			// Tier 3
			{Name: "ipgeolocation.io", BaseURL: "https://api.ipgeolocation.io/ipgeo", Priority: 3, Token: "", Enabled: true, Timeout: 5 * time.Second},
			{Name: "ipwho.is", BaseURL: "https://ipwho.is", Priority: 3, Enabled: true, Timeout: 5 * time.Second},
			{Name: "ipstack.com", BaseURL: "http://api.ipstack.com", Priority: 3, Token: "", Enabled: false, Timeout: 5 * time.Second},
			// Tier 4 — IP-only fallback
			{Name: "api.ipify.org", BaseURL: "https://api.ipify.org", Priority: 4, Enabled: true, Timeout: 3 * time.Second},
			{Name: "icanhazip.com", BaseURL: "https://icanhazip.com", Priority: 4, Enabled: true, Timeout: 3 * time.Second},
			{Name: "checkip.amazonaws.com", BaseURL: "https://checkip.amazonaws.com", Priority: 4, Enabled: true, Timeout: 3 * time.Second},
		},
	}
}

// WriteDefault writes the default config YAML to path, creating parent dirs.
func WriteDefault(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling default config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
