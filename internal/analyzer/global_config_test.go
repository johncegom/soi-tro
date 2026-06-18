package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGlobalConfigPaths(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	is := assert.New(t)

	schemaPath, err := GetSchemaPath()
	is.NoError(err)
	is.Equal(filepath.Join(mockHome, ".config", "soi-tro", "schema.json"), schemaPath)

	configPath, err := GetGlobalConfigPath()
	is.NoError(err)
	is.Equal(filepath.Join(mockHome, ".config", "soi-tro", "config.json"), configPath)
}

func TestSaveAndLoadGlobalAPIKey(t *testing.T) {
	_ = mockUserHomeDir(t)
	is := assert.New(t)

	testKey := "AIzaSyTestAPIKey123456"

	// 1. Lưu API Key
	err := SaveGlobalAPIKey(testKey)
	is.NoError(err)

	// Kiểm tra phân quyền tệp (chỉ cho phép owner đọc/ghi: 0o600)
	configPath, err := GetGlobalConfigPath()
	is.NoError(err)
	info, err := os.Stat(configPath)
	is.NoError(err)
	
	// Trên hệ điều hành Unix, kiểm tra phân quyền (chạy trên Windows sẽ bỏ qua kiểm tra mode này)
	if os.PathSeparator == '/' {
		is.Equal(os.FileMode(0o600), info.Mode().Perm())
	}

	// 2. Tải API Key lên
	loadedKey, err := LoadGlobalAPIKey()
	is.NoError(err)
	is.Equal(testKey, loadedKey)
}

func TestEnsureSchemaFile(t *testing.T) {
	_ = mockUserHomeDir(t)
	is := assert.New(t)
	must := require.New(t)

	// Gán byte schema mặc định giả lập
	defaultSchemaBytes = []byte(`{"type": "OBJECT", "properties": {}}`)

	// 1. Chạy lần đầu -> Phải tạo mới và ký thành công
	schemaPath, err := EnsureSchemaFile()
	is.NoError(err)
	is.FileExists(schemaPath)

	// Đọc và xác thực tính hợp lệ chữ ký
	data, err := os.ReadFile(schemaPath)
	must.NoError(err)
	is.NoError(VerifySchema(data), "Schema được khởi tạo tự động phải có chữ ký hợp lệ")

	// 2. Chạy lần 2 -> Không báo lỗi và giữ nguyên
	schemaPath2, err := EnsureSchemaFile()
	is.NoError(err)
	is.Equal(schemaPath, schemaPath2)
}
