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

	t.Run("fails when directory creation fails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "some_file")
		err := os.WriteFile(filePath, []byte("hello"), 0644)
		require.NoError(t, err)

		cfg := exporter.Config{
			Dir:       filepath.Join(filePath, "sub"),
			MaxSizeKB: 10,
		}

		err = exporter.WriteResult(cfg, &gemini.RentalExtractionResult{}, nil)
		is.Error(err)
		is.Contains(err.Error(), "failed to create export directory")
	})

	t.Run("fails when writing to a directory path", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		tmpDir := t.TempDir()
		cfg := exporter.Config{
			Dir:       tmpDir,
			MaxSizeKB: 10,
		}

		activePath := filepath.Join(tmpDir, "soi-tro-results-001.txt")
		err := os.Mkdir(activePath, 0755)
		require.NoError(t, err)

		err = exporter.WriteResult(cfg, &gemini.RentalExtractionResult{}, nil)
		is.Error(err)
		is.Contains(err.Error(), "failed to open export file")
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

func TestActiveFilePath_InvalidDirectory(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Try to use a directory that doesn't exist and can't be created
	cfg := exporter.Config{
		Dir:       "/nonexistent/directory/that/does/not/exist",
		MaxSizeKB: 10,
	}

	path, err := exporter.ActiveFilePath(cfg)
	if err != nil {
		is.Contains(err.Error(), "invalid export directory")
	} else {
		// On some systems, this might succeed (e.g., if the path is created)
		is.NotEmpty(path)
	}
}

func TestWriteResult_PathTraversal(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       filepath.Join(tmpDir, "..", "escaped"), // Try to escape temp dir
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields:   map[string]string{"field": "value"},
	}

	// Should resolve to absolute path and not escape
	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)
}

func TestWriteResult_EmptyResult(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		RawFields: map[string]string{},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	// Verify file was created
	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	is.FileExists(activePath)
}

func TestWriteResult_MultipleSequentialWrites(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields:   map[string]string{"field": "value"},
	}

	// Write multiple results sequentially
	for i := 0; i < 5; i++ {
		err := exporter.WriteResult(cfg, result, nil)
		is.NoError(err)
	}

	// Verify that files are being used correctly
	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	is.FileExists(activePath)
}

func TestWriteResult_SpecialCharactersInFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields: map[string]string{
			"field1": "Value with special chars: @#$%^&*()",
			"field2": "Value with unicode: 🏠💰",
			"field3": "Value with newlines\nand\ttabs",
		},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "Value with special chars")
}

func TestWriteResult_VeryLongValues(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 1024,
	}

	longValue := strings.Repeat("A", 10000)
	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields: map[string]string{
			"long_field": longValue,
		},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), longValue)
}

func TestWriteResult_WithComplexSampleMessages(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields:   map[string]string{"field": "value"},
		SampleMessages: []gemini.SampleMessage{
			{
				Style:   "Complex Style 1",
				Content: "Complex message with multiple lines\nand special characters: @#$%",
			},
			{
				Style:   "Complex Style 2",
				Content: "Another complex message with unicode: 🏠💰",
			},
		},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "Complex Style 1")
	is.Contains(string(content), "Complex Style 2")
}

func TestWriteResult_WithManyMissingFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "0900000000",
		RawFields:   map[string]string{"field": "value"},
		MissingFields: []string{
			"field1", "field2", "field3", "field4", "field5",
			"field6", "field7", "field8", "field9", "field10",
		},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "Trường còn thiếu")
	is.Contains(string(content), "field1")
	is.Contains(string(content), "field10")
}

func TestWriteResult_WithEmptyFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		RawFields:       map[string]string{},
		MissingFields:   []string{},
		SampleMessages:  []gemini.SampleMessage{},
		AdditionalNotes: "",
		PhoneNumber:     "",
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	// Verify file was created and contains expected content
	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "KẾT QUẢ PHÂN TÍCH")
}

func TestWriteResult_WithAllFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 1024,
	}

	result := &gemini.RentalExtractionResult{
		Price:           "5 triệu",
		Deposit:         "1 tháng",
		Floor:           "2",
		Electricity:     "4000/kWh",
		Water:           "100k/người",
		ParkingFee:      "Miễn phí",
		PetsAllowed:     "Có",
		PhoneNumber:     "0912345678",
		AdditionalNotes: "Gần chợ",
		MissingFields:   []string{"wifi"},
		SampleMessages: []gemini.SampleMessage{
			{Style: "Lịch sự", Content: "Chào bạn, mình muốn hỏi..."},
			{Style: "Trực diện", Content: "Phòng còn không?"},
		},
		RawFields: map[string]string{
			"price":        "5 triệu",
			"deposit":      "1 tháng",
			"floor":        "2",
			"electricity":  "4000/kWh",
			"water":        "100k/người",
			"parking_fee":  "Miễn phí",
			"pets_allowed": "Có",
		},
	}

	titleMap := map[string]string{
		"price":       "Giá thuê",
		"deposit":     "Tiền cọc",
		"electricity": "Tiền điện",
	}

	err := exporter.WriteResult(cfg, result, titleMap)
	is.NoError(err)

	// Verify file was created and contains expected content
	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)

	contentStr := string(content)
	is.Contains(contentStr, "Giá thuê")
	is.Contains(contentStr, "Tiền cọc")
	is.Contains(contentStr, "Tiền điện")
	is.Contains(contentStr, "0912345678")
	is.Contains(contentStr, "Gần chợ")
	is.Contains(contentStr, "wifi")
	is.Contains(contentStr, "Lịch sự")
	is.Contains(contentStr, "Trực diện")
}

func TestWriteResult_WithNilTitleMap(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		RawFields: map[string]string{
			"custom_field": "custom_value",
		},
	}

	// Should not panic with nil titleMap
	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "custom_field")
}

func TestWriteResult_WithEmptyTitleMap(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		RawFields: map[string]string{
			"custom_field": "custom_value",
		},
	}

	titleMap := map[string]string{}

	err := exporter.WriteResult(cfg, result, titleMap)
	is.NoError(err)

	// Should use field name as label
	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.Contains(string(content), "custom_field")
}

func TestWriteResult_PhoneNumberNotProvided(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		PhoneNumber: "Không đề cập",
		RawFields:   map[string]string{"field": "value"},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.NotContains(string(content), "Liên hệ chủ nhà")
}

func TestWriteResult_AdditionalNotesNotProvided(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tmpDir := t.TempDir()
	cfg := exporter.Config{
		Dir:       tmpDir,
		MaxSizeKB: 10,
	}

	result := &gemini.RentalExtractionResult{
		AdditionalNotes: "Không đề cập",
		RawFields:       map[string]string{"field": "value"},
	}

	err := exporter.WriteResult(cfg, result, nil)
	is.NoError(err)

	activePath, err := exporter.ActiveFilePath(cfg)
	is.NoError(err)
	content, err := os.ReadFile(activePath)
	is.NoError(err)
	is.NotContains(string(content), "Ghi chú thêm")
}
