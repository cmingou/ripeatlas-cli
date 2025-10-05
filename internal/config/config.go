package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	APIKey string
}

// LoadWithPriority loads API key with the following priority:
// 1. Environment variable RIPE_ATLAS_API
// 2. ~/.env.key file
// 3. Custom config file (if configPath is provided)
func LoadWithPriority(configPath string) (*Config, error) {
	cfg := &Config{}

	// Priority 1: Check environment variable RIPE_ATLAS_API
	if apiKey := os.Getenv("RIPE_ATLAS_API"); apiKey != "" {
		cfg.APIKey = apiKey
		return cfg, nil
	}

	// Priority 2: Check ~/.env.key
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfigPath := filepath.Join(homeDir, ".env.key")
		if _, err := os.Stat(homeConfigPath); err == nil {
			cfg, err = loadFromFile(homeConfigPath)
			if err == nil && cfg.APIKey != "" {
				return cfg, nil
			}
		}
	}

	// Priority 3: Check custom config file (if provided)
	if configPath != "" {
		cfg, err = loadFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		if cfg.APIKey != "" {
			return cfg, nil
		}
	}

	return nil, fmt.Errorf("RIPE Atlas API key not found. Please set RIPE_ATLAS_API environment variable or create ~/.env.key file")
}

// Load reads configuration from env.key file (kept for backward compatibility)
func Load(filePath string) (*Config, error) {
	return loadFromFile(filePath)
}

// loadFromFile reads configuration from a file
func loadFromFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	cfg := &Config{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		// Support both RIPE_ATLAS_API and RIPE_ATLAS_KEY
		if key == "RIPE_ATLAS_API" || key == "RIPE_ATLAS_KEY" {
			cfg.APIKey = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("RIPE_ATLAS_API or RIPE_ATLAS_KEY not found in config file")
	}

	return cfg, nil
}
