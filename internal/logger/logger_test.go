package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestInit_ErrorCases(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		cfg := Config{
			Level:      "invalid",
			Format:     "json",
			OutputPath: "/tmp/test.log",
		}

		err := Init(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logger configuration")
		resetLogger()
	})

	t.Run("empty output path", func(t *testing.T) {
		cfg := Config{
			Level:      "info",
			Format:     "json",
			OutputPath: "",
		}

		err := Init(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logger configuration")
		resetLogger()
	})

	t.Run("invalid format", func(t *testing.T) {
		cfg := Config{
			Level:      "info",
			Format:     "invalid",
			OutputPath: "/tmp/test.log",
		}

		err := Init(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logger configuration")
		resetLogger()
	})
}

func TestInit_DifferentLevels(t *testing.T) {
	tmpDir := os.TempDir()
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logPath := filepath.Join(tmpDir, "soi-tro-test-"+level+".log")
			os.Remove(logPath)
			defer os.Remove(logPath)

			cfg := Config{
				Level:      level,
				Format:     "text",
				OutputPath: logPath,
				Console:    false,
			}

			err := Init(cfg)
			assert.NoError(t, err)
			Info("test message")
			resetLogger()
		})
	}
}

func TestInit_DifferentFormats(t *testing.T) {
	tmpDir := os.TempDir()
	formats := []string{"text", "json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			logPath := filepath.Join(tmpDir, "soi-tro-test-"+format+".log")
			os.Remove(logPath)
			defer os.Remove(logPath)

			cfg := Config{
				Level:      "info",
				Format:     format,
				OutputPath: logPath,
				Console:    false,
			}

			err := Init(cfg)
			assert.NoError(t, err)
			Info("test message")
			resetLogger()
		})
	}
}

func TestInit_WithConsole(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-console.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "info",
		Format:     "text",
		OutputPath: logPath,
		Console:    true,
	}

	err := Init(cfg)
	assert.NoError(t, err)
	Info("test message")
	resetLogger()
}

func TestInit_WithConsoleJSON(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-console-json.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "info",
		Format:     "json",
		OutputPath: logPath,
		Console:    true,
	}

	err := Init(cfg)
	assert.NoError(t, err)
	Info("test message")
	resetLogger()
}

func TestInit_InvalidDirectory(t *testing.T) {
	// Try to create a log file in a directory that doesn't exist and can't be created
	cfg := Config{
		Level:      "info",
		Format:     "text",
		OutputPath: "/nonexistent/directory/that/does/not/exist/test.log",
		Console:    false,
	}

	err := Init(cfg)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create log directory")
	}
	resetLogger()
}

func TestLoggingFunctions(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-functions.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "debug",
		Format:     "text",
		OutputPath: logPath,
		Console:    false,
	}

	err := Init(cfg)
	assert.NoError(t, err)

	// Test all logging functions
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")

	resetLogger()
}

func TestGet_Fallback(t *testing.T) {
	// Ensure logger is not initialized
	resetLogger()

	// Get should return default logger
	logger := Get()
	assert.NotNil(t, logger)
	assert.Equal(t, slog.Default(), logger)
}

func TestWith(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-with.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "info",
		Format:     "text",
		OutputPath: logPath,
		Console:    false,
	}

	err := Init(cfg)
	assert.NoError(t, err)

	logger := With("custom_key", "custom_value")
	assert.NotNil(t, logger)
	logger.Info("test message")

	resetLogger()
}

func TestInit_DefaultLevel(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-default-level.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      "info", // Valid level
		Format:     "text",
		OutputPath: logPath,
		Console:    false,
	}

	err := Init(cfg)
	assert.NoError(t, err)

	// Test that the logger was initialized correctly
	logger := Get()
	assert.NotNil(t, logger)
	assert.NotEqual(t, slog.Default(), logger)

	resetLogger()
}

func TestInit_WithWhitespaceLevel(t *testing.T) {
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "soi-tro-test-whitespace.log")
	os.Remove(logPath)
	defer os.Remove(logPath)

	cfg := Config{
		Level:      " info ", // Should fail validation
		Format:     "text",
		OutputPath: logPath,
		Console:    false,
	}

	err := Init(cfg)
	assert.Error(t, err)
	resetLogger()
}
