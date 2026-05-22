package exporter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"soi-tro/internal/exporter"
	"soi-tro/internal/gemini"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	cfg := exporter.DefaultConfig()

	is.Equal(".", cfg.Dir, "default directory should be '.'")
	is.Equal(1024, cfg.MaxSizeKB, "default max size should be 1024 KB")
}

func TestWriteResultAndActiveFilePath(t *testing.T) {
	t.Parallel()

	t.Run("successful write and path resolution", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		tmpDir := t.TempDir()
		cfg := exporter.Config{
			Dir:       tmpDir,
			MaxSizeKB: 10,
		}

		result := &gemini.RentalExtractionResult{
			Price:           "3.5 triệu/tháng",
			Deposit:         "3.5 triệu",
			PhoneNumber:     "0987654321",
			AdditionalNotes: "Gần ngã tư",
			MissingFields:   []string{"parking_fee"},
			SampleMessages: []gemini.SampleMessage{
				{Style: "Lịch sự", Content: "Chào anh/chị, tôi muốn hỏi thuê..."},
			},
			RawFields: map[string]string{
				"price":   "3.5 triệu/tháng",
				"deposit": "3.5 triệu",
			},
		}

		titleMap := map[string]string{
			"price":   "Giá thuê",
			"deposit": "Đặt cọc",
		}

		// 1. Write the result
		err := exporter.WriteResult(cfg, result, titleMap)
		must.NoError(err, "WriteResult should succeed")

		// 2. Retrieve active file path
		activePath, err := exporter.ActiveFilePath(cfg)
		must.NoError(err, "ActiveFilePath should succeed")

		is.True(strings.HasSuffix(activePath, "soi-tro-results-001.txt"), "expected active path to end with 'soi-tro-results-001.txt', got %q", activePath)

		// 3. Read back the file content and verify formatting
		bytes, err := os.ReadFile(activePath)
		must.NoError(err, "failed to read exported file")

		content := string(bytes)
		expectedSubstrings := []string{
			"KẾT QUẢ PHÂN TÍCH",
			"Liên hệ chủ nhà",
			"0987654321",
			"Giá thuê",
			"3.5 triệu/tháng",
			"Ghi chú thêm:",
			"Gần ngã tư",
			"Trường còn thiếu:",
			"parking_fee",
			"Tin nhắn mẫu:",
			"[Lịch sự]",
		}

		for _, sub := range expectedSubstrings {
			is.Contains(content, sub, "exported content should contain %q", sub)
		}
	})

	t.Run("file rollover when max size reached", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		tmpDir := t.TempDir()
		// Setup config with tiny max size (1 KB)
		cfg := exporter.Config{
			Dir:       tmpDir,
			MaxSizeKB: 1, // 1 KB = 1024 bytes
		}

		result := &gemini.RentalExtractionResult{
			PhoneNumber: "0900000000",
			RawFields: map[string]string{
				"field": strings.Repeat("A", 1100), // Ensure formatted size exceeds 1KB
			},
		}

		// First write (creates file 1)
		err := exporter.WriteResult(cfg, result, nil)
		must.NoError(err, "first write should succeed")

		// Active path should now be file 2 (since file 1 is full)
		pathNext1, err := exporter.ActiveFilePath(cfg)
		must.NoError(err, "ActiveFilePath should succeed after first write")
		is.True(strings.HasSuffix(pathNext1, "soi-tro-results-002.txt"), "expected active path after first write to be file 2, got %q", pathNext1)

		// Second write (creates file 2)
		err = exporter.WriteResult(cfg, result, nil)
		must.NoError(err, "second write should succeed")

		// Active path should now be file 3 (since both file 1 and 2 are full)
		pathNext2, err := exporter.ActiveFilePath(cfg)
		must.NoError(err, "ActiveFilePath should succeed after second write")
		is.True(strings.HasSuffix(pathNext2, "soi-tro-results-003.txt"), "expected active path after second write to be file 3, got %q", pathNext2)

		// Ensure both created files exist in the sandbox
		file1Path := filepath.Join(tmpDir, "soi-tro-results-001.txt")
		file2Path := filepath.Join(tmpDir, "soi-tro-results-002.txt")

		is.FileExists(file1Path, "file 1 should exist at %q", file1Path)
		is.FileExists(file2Path, "file 2 should exist at %q", file2Path)
	})
}

// ExampleDefaultConfig shows how to fetch the sensible default exporter configurations.
func ExampleDefaultConfig() {
	cfg := exporter.DefaultConfig()
	fmt.Println(cfg.Dir)
	fmt.Println(cfg.MaxSizeKB)

	// Output:
	// .
	// 1024
}
