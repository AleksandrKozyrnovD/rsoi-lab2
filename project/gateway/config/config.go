package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP HTTPConfig     `yaml:"http"`
	URLs ServicesConfig `yaml:"url"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type ServicesConfig struct {
	Reservation string `yaml:"reservation"`
	Loyalty     string `yaml:"loyalty"`
	Payment     string `yaml:"payment"`
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
