// Package exporter appends analyzed rentals to rolling text files.
package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"soi-tro/internal/gemini"
	"soi-tro/internal/logger"
	"strings"
	"time"
)

const (
	filePrefix    = "soi-tro-results"
	defaultMaxKB  = 1024 // 1 MB
	maxFileNumber = 9999
	separator     = "\n================================================================================\n"
)

// Config holds export configuration.
type Config struct {
	Dir       string `json:"dir"`
	MaxSizeKB int    `json:"max_size_kb"`
}

// DefaultConfig returns sensible defaults pointing to the current working directory.
func DefaultConfig() Config {
	return Config{
		Dir:       ".",
		MaxSizeKB: defaultMaxKB,
	}
}

// WriteResult appends a formatted result to the current active export file.
// It automatically rolls over to a new numbered file when the size limit is reached.
// All security fixes are applied:
//   - filepath.Abs() prevents path traversal from user-supplied dir
//   - File permissions are 0o600 (owner read/write only) to protect PII
//   - File numbering is capped at maxFileNumber to prevent infinite loops
//   - Each file handle uses a dedicated scoped helper to ensure safe defer/close
//
//nolint:wsl_v5 // Preserve the existing export flow around logging stages.
func WriteResult(cfg Config, result *gemini.RentalExtractionResult, titleMap map[string]string) (err error) {
	started := time.Now()
	failureStage := "resolve_directory"
	log := logger.With("operation", "export.write_result")
	defer func() { logger.LogOperationResult(log, started, failureStage, err) }()

	// Security: resolve to absolute path to neutralize any ".." path traversal
	// in the user-supplied directory before any filesystem operation.
	absDir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return fmt.Errorf("invalid export directory: %w", err)
	}
	cfg.Dir = absDir

	failureStage = "create_directory"
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	maxBytes := int64(cfg.MaxSizeKB) * 1024
	content := formatResult(result, titleMap)

	failureStage = "find_active_file"
	activeFile, num, err := findActiveFile(cfg.Dir, maxBytes)
	if err != nil {
		return fmt.Errorf("failed to determine active export file: %w", err)
	}

	// Check if content fits in the active file; if not, roll over.
	info, statErr := os.Stat(activeFile)
	if statErr == nil && info.Size() > 0 && info.Size()+int64(len(content)) > maxBytes {
		num++
		if num > maxFileNumber {
			failureStage = "file_limit"
			return fmt.Errorf("reached maximum export file limit (%d)", maxFileNumber)
		}
		activeFile = buildFilePath(cfg.Dir, num)
	}

	failureStage = "append_file"
	return appendToFile(activeFile, content)
}

// ActiveFilePath returns the path of the currently active export file for display purposes.
//
//nolint:wsl_v5 // Preserve the existing export flow around logging stages.
func ActiveFilePath(cfg Config) (path string, err error) {
	started := time.Now()
	failureStage := "resolve_directory"
	log := logger.With("operation", "export.active_file_path")
	defer func() { logger.LogOperationResult(log, started, failureStage, err) }()

	// Security: resolve absolute path consistently with WriteResult.
	absDir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return "", fmt.Errorf("invalid export directory: %w", err)
	}
	cfg.Dir = absDir

	maxBytes := int64(cfg.MaxSizeKB) * 1024
	failureStage = "find_active_file"
	path, _, err = findActiveFile(cfg.Dir, maxBytes)
	return path, err
}

// appendToFile opens a file in append mode and writes content to it.
// Each call gets its own scoped file handle; the Close error is returned so a
// failed flush is not silently lost.
// Security: uses 0o600 permissions — owner read/write only — to protect PII.
func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open export file %q: %w", path, err)
	}

	if _, err := fmt.Fprint(f, content); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write to export file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close export file %q: %w", path, err)
	}
	return nil
}

// findActiveFile scans the export directory for the lowest-numbered file that
// still has capacity. Returns the path and its sequence number.
// Security: capped at maxFileNumber to prevent an infinite loop if every
// file somehow always reports as full.
func findActiveFile(dir string, maxBytes int64) (string, int, error) {
	for num := 1; num <= maxFileNumber; num++ {
		path := buildFilePath(dir, num)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			// This slot does not exist yet — it becomes the new active file.
			return path, num, nil
		}
		if err != nil {
			return "", 0, fmt.Errorf("failed to stat export file %q: %w", path, err)
		}
		if info.Size() < maxBytes {
			return path, num, nil
		}
	}
	return "", 0, fmt.Errorf("reached maximum export file limit (%d)", maxFileNumber)
}

func buildFilePath(dir string, num int) string {
	return filepath.Join(dir, fmt.Sprintf("%s-%03d.txt", filePrefix, num))
}

// formatResult renders a RentalExtractionResult as plain, copy-pasteable text.
// AI-generated content is written verbatim — safe for plain-text format with
// no injection risk (no SQL, no HTML, no shell context).
func formatResult(result *gemini.RentalExtractionResult, titleMap map[string]string) string {
	var b strings.Builder

	b.WriteString(separator)
	fmt.Fprintf(&b, "KẾT QUẢ PHÂN TÍCH  —  %s\n", time.Now().Format("02/01/2006 15:04:05"))
	b.WriteString("================================================================================\n\n")

	if result.PhoneNumber != "" && result.PhoneNumber != "Không đề cập" {
		fmt.Fprintf(&b, "%-26s: %s\n", "Liên hệ chủ nhà", result.PhoneNumber)
	}

	// Write all extracted structured fields in a consistent, aligned format.
	for k, v := range result.RawFields {
		if k == "phone_number" || k == "additional_notes" {
			continue
		}
		label := k
		if t, ok := titleMap[k]; ok && t != "" {
			label = t
		}
		fmt.Fprintf(&b, "%-26s: %s\n", label, v)
	}

	if result.AdditionalNotes != "" && result.AdditionalNotes != "Không đề cập" {
		fmt.Fprintf(&b, "\nGhi chú thêm:\n  %s\n", result.AdditionalNotes)
	}

	if len(result.MissingFields) > 0 {
		b.WriteString("\nTrường còn thiếu:\n")
		for _, m := range result.MissingFields {
			fmt.Fprintf(&b, "  - %s\n", m)
		}
	}

	if len(result.SampleMessages) > 0 {
		b.WriteString("\nTin nhắn mẫu:\n")
		for _, msg := range result.SampleMessages {
			fmt.Fprintf(&b, "\n  [%s]\n  %s\n", msg.Style, msg.Content)
		}
	}

	b.WriteString("\n")
	return b.String()
}
