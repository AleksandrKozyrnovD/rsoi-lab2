package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
}

// type LoggerConfig struct {
// 	Level string
// 	File  string
// }

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Driver   string         `yaml:"driver"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func InitConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}
