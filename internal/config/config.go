package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	APIKey string
}

// Load reads configuration from env.key file
func Load(filePath string) (*Config, error) {
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

		if key == "RIPE_ATLAS_KEY" {
			cfg.APIKey = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("RIPE_ATLAS_KEY not found in config file")
	}

	return cfg, nil
}
