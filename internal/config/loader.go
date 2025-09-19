// internal/config/loader.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Apply defaults (schema fallback logic)
	applyDefaults(&cfg)

	return &cfg, nil
}

// applyDefaults fills in any default values that should be set
// when they are omitted in the YAML.
func applyDefaults(cfg *Config) {
	for i := range cfg.Tasks {
		task := &cfg.Tasks[i]

		// default schema prefix if not provided
		if task.Table != "" && !containsDot(task.Table) {
			task.Table = "public." + task.Table
		}

		// default structure if not set
		if task.Structure == "" {
			task.Structure = "stream"
		}
	}
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}
