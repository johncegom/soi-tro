package logger

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds the configuration for the logger
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	OutputPath string // path to log file
	Console    bool   // enable console output
	
	// Rotation settings
	MaxSizeMB  int // maximum size in megabytes before rotation
	MaxBackups int // maximum number of old log files to retain
	MaxAgeDays int // maximum number of days to retain old log files
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	return Config{
		Level:      "info",
		Format:     "text", // text for development, json for production
		OutputPath: filepath.Join(homeDir, ".config", "soi-tro", "app.log"),
		Console:    true,
		MaxSizeMB:  100,
		MaxBackups: 5,
		MaxAgeDays: 30,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() Config {
	cfg := DefaultConfig()
	
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Level = strings.ToLower(level)
	}
	
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Format = strings.ToLower(format)
	}
	
	if path := os.Getenv("LOG_FILE"); path != "" {
		cfg.OutputPath = path
	}
	
	if console := os.Getenv("LOG_CONSOLE"); console != "" {
		cfg.Console = strings.ToLower(console) == "true"
	}
	
	return cfg
}

// IsValid checks if the configuration is valid
func (c Config) IsValid() bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	
	return validLevels[c.Level] && validFormats[c.Format] && c.OutputPath != ""
}
