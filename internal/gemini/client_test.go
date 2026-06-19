package gemini

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func TestNewClient_NoAPIKey(t *testing.T) {
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	c, err := NewClient(context.Background())
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "GEMINI_API_KEY environment variable is not set")
}

func TestNewClient_Success(t *testing.T) {
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "some-mock-key")
	t.Cleanup(func() {
		os.Setenv("GEMINI_API_KEY", origKey)
	})

	c, err := NewClient(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestLoadSchema_FileNotExist(t *testing.T) {
	_, err := LoadSchema("non_existent_file_path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read schema file")
}

func TestLoadSchema_InvalidSignature(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err := os.WriteFile(schemaPath, []byte(`{"type": "OBJECT"}`), 0600)
	assert.NoError(t, err)

	_, err = LoadSchema(schemaPath)
	assert.Error(t, err)
}

func TestLoadSchema_InvalidSchemaStructure(t *testing.T) {
	invalidSchema := []byte(`{"type": "OBJECT", "properties": "invalid_type"}`)
	signedBytes, err := analyzer.SignSchema(invalidSchema)
	require.NoError(t, err)

	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	_, err = LoadSchema(schemaPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode schema JSON")
}

func TestLoadExportConfig_FileNotExist(t *testing.T) {
	_, _, err := LoadExportConfig("non_existent_file_path")
	assert.Error(t, err)
}

func TestLoadExportConfig_InvalidSignature(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err := os.WriteFile(schemaPath, []byte(`{"type": "OBJECT"}`), 0600)
	assert.NoError(t, err)

	_, _, err = LoadExportConfig(schemaPath)
	assert.Error(t, err)
}


func TestLoadExportConfig_InvalidConfigType(t *testing.T) {
	invalidConfigJSON := []byte(`{"type": "OBJECT", "x_export_config": "should_be_object"}`)
	signedBytes, err := analyzer.SignSchema(invalidConfigJSON)
	require.NoError(t, err)

	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	_, _, err = LoadExportConfig(schemaPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse export config")
}

func TestSaveExportConfig_FileNotExist(t *testing.T) {
	err := SaveExportConfig("non_existent_file_path", "/dir", 1024)
	assert.Error(t, err)
}

func TestSaveExportConfig_Untrusted(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err := os.WriteFile(schemaPath, []byte(`{"type": "OBJECT"}`), 0600)
	assert.NoError(t, err)

	err = SaveExportConfig(schemaPath, "/dir", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing schema file is untrusted")
}


func TestExtractRentalInfo_EmptyInputs(t *testing.T) {
	c := &Client{}
	_, err := c.ExtractRentalInfo(context.Background(), "", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either text listing or image file must be provided")
}

func TestExtractRentalInfo_SchemaLoadFail(t *testing.T) {
	_ = mockUserHomeDir(t)
	c := &Client{}
	_, err := c.ExtractRentalInfo(context.Background(), "some listing", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load output schema")
}

func TestExtractRentalInfo_GetSchemaPathFail(t *testing.T) {
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

	c := &Client{}
	_, err := c.ExtractRentalInfo(context.Background(), "some listing", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get schema path")
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newMockClient(t *testing.T, handler roundTripFunc) *Client {
	mockHTTPClient := &http.Client{
		Transport: handler,
	}
	config := &genai.ClientConfig{
		APIKey:     "mock-api-key-123",
		HTTPClient: mockHTTPClient,
	}
	gc, err := genai.NewClient(context.Background(), config)
	require.NoError(t, err)
	return &Client{
		genaiClient: gc,
	}
}

func TestSaveSchema_WriteFileFail(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	dirPath := filepath.Join(mockHome, "schema.json")
	err := os.Mkdir(dirPath, 0755)
	require.NoError(t, err)

	err = SaveSchema(dirPath, &genai.Schema{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write schema file")
}

func TestSaveSchema_ExistingUntrusted(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	err := os.WriteFile(schemaPath, []byte(`{"type": "OBJECT", "x_signature": "invalid"}`), 0600)
	require.NoError(t, err)

	err = SaveSchema(schemaPath, &genai.Schema{Type: genai.TypeObject}, nil)
	assert.NoError(t, err)
}

func TestExtractRentalInfo_Success(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")

	baseJSON := `{
		"type": "OBJECT",
		"properties": {
			"price": {
				"type": "STRING",
				"title": "Giá thuê"
			},
			"missing_fields": {
				"type": "ARRAY",
				"items": {
					"type": "STRING"
				}
			}
		},
		"x_required_fields": ["price"]
	}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	apiResponse := `{
		"candidates": [
			{
				"content": {
					"parts": [
						{
							"text": "{\n  \"price\": \"3,500,000 VND\",\n  \"deposit\": \"3,500,000 VND\",\n  \"phone_number\": \"0987654321\",\n  \"additional_notes\": \"Gần ngã tư\",\n  \"missing_fields\": [\"parking_fee\"],\n  \"sample_messages\": [\n    {\n      \"style\": \"Lịch sự\",\n      \"content\": \"Chào anh/chị...\"\n    }\n  ]\n}"
						}
					]
				}
			}
		]
	}`

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "generateContent")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
			Header:     make(http.Header),
		}, nil
	}))

	res, err := c.ExtractRentalInfo(context.Background(), "Phòng trọ 3.5tr", nil, "", []string{"parking_fee"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "3,500,000 VND", res.Price)
	assert.Equal(t, "0987654321", res.PhoneNumber)
	assert.Contains(t, res.MissingFields, "parking_fee")
	assert.Len(t, res.SampleMessages, 1)
	assert.Equal(t, "3,500,000 VND", res.RawFields["price"])
}

func TestExtractRentalInfo_SuccessImage(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")
	baseJSON := `{"type": "OBJECT", "properties": {"missing_fields": {"type": "ARRAY", "items": {"type": "STRING"}}}}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	apiResponse := `{"candidates": [{"content": {"parts": [{"text": "{\"price\": \"4,000,000 VND\", \"custom_number_field\": 123}"}]}}]}`

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
			Header:     make(http.Header),
		}, nil
	}))

	res, err := c.ExtractRentalInfo(context.Background(), "some listing", []byte("fake-image-bytes"), "image/png", nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "4,000,000 VND", res.Price)
	assert.Equal(t, "123", res.RawFields["custom_number_field"])
}

func TestExtractRentalInfo_APIFail(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")
	baseJSON := `{"type": "OBJECT", "properties": {}}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("internal server error")),
			Header:     make(http.Header),
		}, nil
	}))

	_, err = c.ExtractRentalInfo(context.Background(), "Phòng trọ", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate content from Gemini API")
}

func TestExtractRentalInfo_NoCandidates(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")
	baseJSON := `{"type": "OBJECT", "properties": {}}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	apiResponse := `{"candidates": []}`

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
			Header:     make(http.Header),
		}, nil
	}))

	_, err = c.ExtractRentalInfo(context.Background(), "Phòng trọ", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no response candidates returned by Gemini")
}

func TestExtractRentalInfo_MalformedJSON(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")
	baseJSON := `{"type": "OBJECT", "properties": {}}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	apiResponse := `{"candidates": [{"content": {"parts": [{"text": "invalid-json-here"}]}}]}`

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
			Header:     make(http.Header),
		}, nil
	}))

	_, err = c.ExtractRentalInfo(context.Background(), "Phòng trọ", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal JSON response from model")
}

func TestLoadExportConfig_NotConfigured(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	baseJSON := `{"type": "OBJECT"}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	cfg, configured, err := LoadExportConfig(schemaPath)
	assert.NoError(t, err)
	assert.False(t, configured)
	assert.Empty(t, cfg.Dir)
}

func TestSaveExportConfig_WriteFileFail(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	baseJSON := `{"type": "OBJECT"}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	err = os.Chmod(schemaPath, 0400)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chmod(schemaPath, 0600)
	})

	err = SaveExportConfig(schemaPath, "/dir", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write schema file")
}

func TestSaveSchema_WriteFileFailReadOnly(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")
	baseJSON := `{"type": "OBJECT"}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	err = os.Chmod(schemaPath, 0400)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chmod(schemaPath, 0600)
	})

	err = SaveSchema(schemaPath, &genai.Schema{Type: genai.TypeObject}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write schema file")
}

func TestSaveSchema_MarshalFail(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	schemaPath := filepath.Join(mockHome, "schema.json")

	cyclicSchema := &genai.Schema{
		Type: genai.TypeObject,
	}
	cyclicSchema.Items = cyclicSchema

	err := SaveSchema(schemaPath, cyclicSchema, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal schema")
}

func TestExtractRentalInfo_NoParts(t *testing.T) {
	mockHome := mockUserHomeDir(t)
	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	schemaPath := filepath.Join(configDir, "schema.json")
	baseJSON := `{"type": "OBJECT", "properties": {}}`
	signedBytes, err := analyzer.SignSchema([]byte(baseJSON))
	require.NoError(t, err)
	err = os.WriteFile(schemaPath, signedBytes, 0600)
	require.NoError(t, err)

	apiResponse := `{"candidates": [{"content": {"parts": []}}]}`

	c := newMockClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
			Header:     make(http.Header),
		}, nil
	}))

	_, err = c.ExtractRentalInfo(context.Background(), "Phòng trọ", nil, "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no response candidates returned by Gemini")
}

