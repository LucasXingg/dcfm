package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	Model   string `json:"model,omitempty"`
}

func getDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userConfigDir, "dcfm"), nil
}

func getPath() (string, error) {
	dir, err := getDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	var cfg Config
	path, err := getPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return empty config if not exists
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Environment variable overrides
	if envKey := os.Getenv("DCFM_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}
	if envURL := os.Getenv("DCFM_BASE_URL"); envURL != "" {
		cfg.BaseURL = envURL
	}
	if envModel := os.Getenv("DCFM_MODEL"); envModel != "" {
		cfg.Model = envModel
	}

	// Set defaults
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}

	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := getDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := getPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
