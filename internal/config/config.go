package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration.
type Config struct {
	Cache          CacheConfig         `yaml:"cache"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Providers      []ProviderConfig    `yaml:"providers"`
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	TTL     time.Duration `yaml:"ttl"`
	MaxSize int           `yaml:"max_size"`
}

// CircuitBreakerConfig holds circuit breaker settings.
type CircuitBreakerConfig struct {
	MaxFailures  int           `yaml:"max_failures"`
	ResetTimeout time.Duration `yaml:"reset_timeout"`
}

// ProviderConfig describes a single IP provider.
type ProviderConfig struct {
	Name     string        `yaml:"name"`
	BaseURL  string        `yaml:"base_url"`
	Priority int           `yaml:"priority"`
	Token    string        `yaml:"token"`
	Enabled  bool          `yaml:"enabled"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Load reads the YAML config from path. If path is empty, it tries the default
// location (~/.ipinfo/config.yaml). If that also does not exist, it generates
// and writes a default config file and returns it.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := DefaultConfig()
		if writeErr := WriteDefault(path, cfg); writeErr != nil {
			// Non-fatal: return in-memory default.
			fmt.Fprintf(os.Stderr, "warning: could not write default config to %s: %v\n", path, writeErr)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
