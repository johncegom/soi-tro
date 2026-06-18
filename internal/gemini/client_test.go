package gemini

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
	"soi-tro/internal/analyzer"
)

// mockUserHomeDir giả lập thư mục Home của người dùng sang thư mục tạm thời trong thời gian chạy test
func mockUserHomeDir(t *testing.T) string {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	return tmpDir
}

func TestSchemaAndExportConfig(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	is := assert.New(t)
	must := require.New(t)

	// Tạo thư mục cấu hình giả lập
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0o700)
	must.NoError(err)

	schemaPath := filepath.Join(configDir, "schema.json")

	// 1. Tạo một JSON Schema hợp lệ để ký
	baseJSON := `{
		"type": "OBJECT",
		"properties": {
			"price": {
				"type": "STRING",
				"title": "Giá thuê"
			}
		},
		"x_required_fields": ["price"],
		"x_signature": ""
	}`

	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	must.NoError(err)
	err = os.WriteFile(schemaPath, signedBytes, 0644)
	must.NoError(err)

	// 2. Kiểm thử LoadSchema
	schema, err := LoadSchema(schemaPath)
	is.NoError(err)
	is.NotNil(schema)
	is.Equal(genai.TypeObject, schema.Type)
	is.Contains(schema.Properties, "price")

	// 3. Kiểm thử SaveSchema (Thêm thuộc tính mới)
	schema.Properties["deposit"] = &genai.Schema{
		Type:  genai.TypeString,
		Title: "Tiền cọc",
	}
	err = SaveSchema(schemaPath, schema, []string{"price", "deposit"})
	is.NoError(err)

	// Tải lại và kiểm tra thuộc tính mới
	loadedSchema, err := LoadSchema(schemaPath)
	is.NoError(err)
	is.Contains(loadedSchema.Properties, "deposit")
	is.Equal("Tiền cọc", loadedSchema.Properties["deposit"].Title)

	// 4. Kiểm thử SaveExportConfig & LoadExportConfig
	err = SaveExportConfig(schemaPath, "/test/export/dir", 2048)
	is.NoError(err)

	exportCfg, configured, err := LoadExportConfig(schemaPath)
	is.NoError(err)
	is.True(configured)
	is.Equal("/test/export/dir", exportCfg.Dir)
	is.Equal(2048, exportCfg.MaxSizeKB)
}
