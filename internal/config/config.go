package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Tasks    []TaskConfig   `yaml:"tasks"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"` // e.g., disable, require
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"` // e.g., 0
}

type TaskConfig struct {
	Name      string            `yaml:"name"` // Required
	Table     string            `yaml:"table"`
	Alias     string            `yaml:"alias,omitempty"`
	Structure string            `yaml:"structure"`
	Key       string            `yaml:"key,omitempty"`
	Value     string            `yaml:"value,omitempty"`
	Score     string            `yaml:"score,omitempty"`
	Fields    []string          `yaml:"fields,omitempty"`
	KeyPrefix string            `yaml:"key_prefix,omitempty"`
	Schedule  string            `yaml:"schedule"`
	ColumnMap map[string]string `yaml:"column_map,omitempty"` // key: alias, value: actual DB field
}

func (t *TaskConfig) ResolveColumn(logicalName string) string {
	if actual, ok := t.ColumnMap[logicalName]; ok {
		return actual
	}
	return logicalName
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &cfg, nil
}
