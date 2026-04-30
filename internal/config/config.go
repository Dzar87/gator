package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	payload, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	cfg := Config{}
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) ToPrettyJson() (string, error) {
	payload, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	return write(c)
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve HOME directory: %w", err)
	}
	configPath := filepath.Join(home, configFileName)
	return configPath, nil
}

func write(cfg *Config) error {
	payload, err := json.MarshalIndent(*cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.WriteFile(configPath, payload, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
