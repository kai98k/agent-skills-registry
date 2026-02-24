package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CLIConfig struct {
	APIURL string `yaml:"api_url" json:"api_url"`
	Token  string `yaml:"token" json:"token"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agentskills", "config.yaml")
}

func Load() (*CLIConfig, error) {
	path := DefaultPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{APIURL: "http://localhost:8000"}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg CLIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:8000"
	}
	return &cfg, nil
}

func Save(cfg *CLIConfig) error {
	path := DefaultPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
