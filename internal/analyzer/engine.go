package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration settings
type Config struct {
	RequiredFields []string `json:"x_required_fields"`
}

// LoadConfig reads the configuration from a JSON file path
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config json: %w", err)
	}

	return &config, nil
}

// SaveConfig writes the configuration back to a JSON file path
func SaveConfig(filePath string, config *Config) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config JSON: %w", err)
	}

	return nil
}
