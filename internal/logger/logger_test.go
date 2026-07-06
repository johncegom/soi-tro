package logger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("Expected default level to be 'info', got '%s'", cfg.Level)
	}

	if cfg.Format != "text" {
		t.Errorf("Expected default format to be 'text', got '%s'", cfg.Format)
	}

	if cfg.Console != true {
		t.Errorf("Expected default console to be true")
	}

	if cfg.MaxSizeMB != 100 {
		t.Errorf("Expected default MaxSizeMB to be 100, got %d", cfg.MaxSizeMB)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_CONSOLE", "false")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("LOG_CONSOLE")
	}()

	cfg := LoadFromEnv()

	if cfg.Level != "debug" {
		t.Errorf("Expected level to be 'debug', got '%s'", cfg.Level)
	}

	if cfg.Format != "json" {
		t.Errorf("Expected format to be 'json', got '%s'", cfg.Format)
	}

	if cfg.Console != false {
		t.Errorf("Expected console to be false")
	}
}

func TestConfigIsValid(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "/tmp/test.log",
			},
			want: true,
		},
		{
			name: "invalid level",
			cfg: Config{
				Level:      "invalid",
				Format:     "json",
				OutputPath: "/tmp/test.log",
			},
			want: false,
		},
		{
			name: "invalid format",
			cfg: Config{
				Level:      "info",
				Format:     "invalid",
				OutputPath: "/tmp/test.log",
			},
			want: false,
		},
		{
			name: "empty path",
			cfg: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsValid(); got != tt.want {
				t.Errorf("Config.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Use a fixed temp directory instead of t.TempDir() to avoid cleanup issues
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test.log")

	// Clean up any existing test file
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "debug",
		Format:     "json",
		OutputPath: logPath,
		Console:    false,
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Test logging
	Info("test message", "key", "value")

	// Reset global logger for other tests
	resetLogger()
}

func TestContext(t *testing.T) {
	ctx := context.Background()

	// Test NewContext
	ctxWithID := NewContext(ctx)
	requestID := GetRequestID(ctxWithID)

	if requestID == "" {
		t.Error("Expected request ID to be set")
	}

	// Test WithRequestID
	customID := "custom-123"
	ctxWithCustomID := WithRequestID(ctx, customID)
	retrievedID := GetRequestID(ctxWithCustomID)

	if retrievedID != customID {
		t.Errorf("Expected request ID to be '%s', got '%s'", customID, retrievedID)
	}
}
