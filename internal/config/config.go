package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Tasks    []TaskConfig   `yaml:"tasks"`
	Server   ServerConfig   `yaml:"server_port"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type TaskConfig struct {
	Name      string            `yaml:"name"`
	Table     string            `yaml:"table"`
	Alias     string            `yaml:"alias,omitempty"`
	Structure string            `yaml:"structure"`
	Key       string            `yaml:"key,omitempty"`
	Value     string            `yaml:"value,omitempty"`
	Score     string            `yaml:"score,omitempty"`
	Fields    []string          `yaml:"fields,omitempty"`
	KeyPrefix string            `yaml:"key_prefix,omitempty"`
	Schedule  string            `yaml:"schedule"`
	ColumnMap map[string]string `yaml:"column_map,omitempty"`
	Tracking  *TrackingConfig   `yaml:"tracking,omitempty"`
}

type TrackingConfig struct {
	Column       string `yaml:"column"`
	Operator     string `yaml:"operator"`       // ">" or "<"
	LastValueKey string `yaml:"last_value_key"` // Redis key to store last seen value
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

func (t *TaskConfig) EffectiveRedisKey() string {
	if t.Alias != "" {
		return t.Alias
	}
	return t.Table
}

func (t *TaskConfig) ResolveColumn(logicalName string) string {
	if actual, ok := t.ColumnMap[logicalName]; ok {
		return actual
	}
	return logicalName
}
