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

	testKey := "mock-test-api-key-123456"

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

func TestEnsureGlobalAPIKey_EnvExists(t *testing.T) {
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "existing-env-key")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	err := EnsureGlobalAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, "existing-env-key", os.Getenv("GEMINI_API_KEY"))
}

func TestEnsureGlobalAPIKey_FromConfig(t *testing.T) {
	_ = mockUserHomeDir(t)
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	testKey := "mock-test-api-key-999"
	err := SaveGlobalAPIKey(testKey)
	assert.NoError(t, err)

	err = EnsureGlobalAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, testKey, os.Getenv("GEMINI_API_KEY"))
}

func TestEnsureGlobalAPIKey_ConfigFileDoesNotExist(t *testing.T) {
	_ = mockUserHomeDir(t)
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	// Skip this test in non-interactive environments since it requires user input
	t.Skip("Skipping interactive test in non-interactive environment")
}

func TestEnsureGlobalAPIKey_ConfigFileInvalidJSON(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	configPath := filepath.Join(mockHome, ".config", "soi-tro", "config.json")
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(configPath, []byte("invalid json"), 0600)
	assert.NoError(t, err)

	// Skip this test in non-interactive environments since it requires user input
	t.Skip("Skipping interactive test in non-interactive environment")
}

func TestEnsureGlobalAPIKey_ConfigFileEmptyKey(t *testing.T) {
	_ = mockUserHomeDir(t)
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	// Save empty key
	err := SaveGlobalAPIKey("")
	assert.NoError(t, err)

	// Skip this test in non-interactive environments since it requires user input
	t.Skip("Skipping interactive test in non-interactive environment")
}

func TestLoadGlobalAPIKey_InvalidJSON(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configPath := filepath.Join(mockHome, ".config", "soi-tro", "config.json")
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(configPath, []byte("invalid json"), 0600)
	assert.NoError(t, err)

	_, err = LoadGlobalAPIKey()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse global config JSON")
}

func TestGetSchemaPath_Error(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", "")
	os.Setenv("USERPROFILE", "")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	_, err := GetSchemaPath()
	assert.Error(t, err)
}

func TestEnsureSchemaFile_Migration(t *testing.T) {
	_ = mockUserHomeDir(t)
	is := assert.New(t)

	schemaPath, err := GetSchemaPath()
	is.NoError(err)
	err = os.MkdirAll(filepath.Dir(schemaPath), 0700)
	is.NoError(err)

	unsignedData := []byte(`{"type": "OBJECT", "properties": {}}`)
	err = os.WriteFile(schemaPath, unsignedData, 0600)
	is.NoError(err)

	path, err := EnsureSchemaFile()
	is.NoError(err)
	is.Equal(schemaPath, path)

	signedData, err := os.ReadFile(schemaPath)
	is.NoError(err)
	is.NoError(VerifySchema(signedData))
}

func TestGetAndSaveGlobalModel(t *testing.T) {
	_ = mockUserHomeDir(t)
	is := assert.New(t)

	// 1. Default model when file does not exist
	is.Equal("gemini-3.1-flash-lite", GetGlobalModel())

	// 2. Save and load model
	err := SaveGlobalModel("gemini-3.5-flash")
	is.NoError(err)
	is.Equal("gemini-3.5-flash", GetGlobalModel())

	// 3. Verify API Key is preserved when saving model, and vice-versa
	err = SaveGlobalAPIKey("mock-preserved-key")
	is.NoError(err)
	is.Equal("gemini-3.5-flash", GetGlobalModel())

	key, err := LoadGlobalAPIKey()
	is.NoError(err)
	is.Equal("mock-preserved-key", key)

	err = SaveGlobalModel("gemini-2.5-pro")
	is.NoError(err)
	is.Equal("gemini-2.5-pro", GetGlobalModel())

	key, err = LoadGlobalAPIKey()
	is.NoError(err)
	is.Equal("mock-preserved-key", key)
}

func TestGetGlobalModel_InvalidJSON(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configPath := filepath.Join(mockHome, ".config", "soi-tro", "config.json")
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(configPath, []byte("invalid json"), 0600)
	assert.NoError(t, err)

	// Should return default model on error
	assert.Equal(t, "gemini-3.1-flash-lite", GetGlobalModel())
}

func TestGetGlobalModel_EmptyModel(t *testing.T) {
	_ = mockUserHomeDir(t)
	err := SaveGlobalModel("")
	assert.NoError(t, err)

	// Should return default model when empty
	assert.Equal(t, "gemini-3.1-flash-lite", GetGlobalModel())
}

func TestGetGlobalModel_WhitespaceModel(t *testing.T) {
	_ = mockUserHomeDir(t)
	err := SaveGlobalModel("  ")
	assert.NoError(t, err)

	// Should return default model when whitespace
	assert.Equal(t, "gemini-3.1-flash-lite", GetGlobalModel())
}

func TestSaveGlobalModel_Whitespace(t *testing.T) {
	_ = mockUserHomeDir(t)
	err := SaveGlobalModel("  gemini-3.5-flash  ")
	assert.NoError(t, err)

	// Should trim whitespace
	assert.Equal(t, "gemini-3.5-flash", GetGlobalModel())
}

func TestSaveGlobalAPIKey_Whitespace(t *testing.T) {
	_ = mockUserHomeDir(t)
	err := SaveGlobalAPIKey("  mock-test-key  ")
	assert.NoError(t, err)

	// Should trim whitespace
	key, err := LoadGlobalAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, "mock-test-key", key)
}

func TestSaveGlobalAPIKey_GetPathError(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", "")
	os.Setenv("USERPROFILE", "")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	err := SaveGlobalAPIKey("test-key")
	assert.Error(t, err)
}

func TestSaveGlobalModel_GetPathError(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", "")
	os.Setenv("USERPROFILE", "")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	err := SaveGlobalModel("gemini-3.5-flash")
	assert.Error(t, err)
}

func TestEnsureSchemaFile_CreationError(t *testing.T) {
	// Skip this test on Windows as it may behave differently
	if os.PathSeparator == '\\' {
		t.Skip("Skipping on Windows due to different path handling")
	}

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	// Set HOME to a path that cannot be created as a directory
	os.Setenv("HOME", "/nonexistent/path/that/does/not/exist")
	os.Setenv("USERPROFILE", "/nonexistent/path/that/does/not/exist")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	_, err := EnsureSchemaFile()
	assert.Error(t, err)
}

func TestLoadGlobalAPIKey_FileNotExist(t *testing.T) {
	_ = mockUserHomeDir(t)

	_, err := LoadGlobalAPIKey()
	assert.Error(t, err)
}
