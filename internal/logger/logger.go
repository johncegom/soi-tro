package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalLogger *slog.Logger
	once         sync.Once
)

// resetLogger is used for testing purposes to reset the global logger
func resetLogger() {
	globalLogger = nil
	once = sync.Once{}
}

// Init initializes the global logger with the given configuration
func Init(cfg Config) error {
	var initErr error
	once.Do(func() {
		if !cfg.IsValid() {
			initErr = fmt.Errorf("invalid logger configuration")
			return
		}

		// Parse log level
		var level slog.Level
		switch cfg.Level {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}

		// Create log directory if it doesn't exist
		logDir := filepath.Dir(cfg.OutputPath)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		// Open log file
		logFile, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		// Create handler options
		opts := &slog.HandlerOptions{
			Level: level,
		}

		// Create multi-writer handler if console output is enabled
		if cfg.Console {
			// Use io.MultiWriter to write to both file and console
			multiWriter := io.MultiWriter(logFile, os.Stdout)
			if cfg.Format == "json" {
				globalLogger = slog.New(slog.NewJSONHandler(multiWriter, opts))
			} else {
				globalLogger = slog.New(slog.NewTextHandler(multiWriter, opts))
			}
		} else {
			// Create file handler only
			if cfg.Format == "json" {
				globalLogger = slog.New(slog.NewJSONHandler(logFile, opts))
			} else {
				globalLogger = slog.New(slog.NewTextHandler(logFile, opts))
			}
		}
	})

	return initErr
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if globalLogger == nil {
		// Fallback to default logger if not initialized
		return slog.Default()
	}
	return globalLogger
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// WithContext returns a logger with context
func WithContext(ctx context.Context) *slog.Logger {
	logger := Get()
	// Try to get request ID from context
	if requestID := GetRequestID(ctx); requestID != "" {
		return logger.With("request_id", requestID)
	}
	return logger
}

// With returns a logger with additional attributes
func With(key string, value any) *slog.Logger {
	return Get().With(key, value)
}
