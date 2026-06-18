package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAndSaveConfig(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "schema.json")

	// 1. Chuẩn bị JSON thô chưa ký
	rawJSON := `{
		"x_required_fields": ["price", "deposit"]
	}`

	// 2. Ghi và cố gắng tải tệp chưa ký -> Phải trả về lỗi ErrSchemaTampered
	err := os.WriteFile(configPath, []byte(rawJSON), 0644)
	must.NoError(err)

	_, err = LoadConfig(configPath)
	is.ErrorIs(err, ErrSchemaTampered, "LoadConfig phải báo lỗi khi schema chưa được ký")

	// 3. Ký nội dung JSON và ghi lại vào tệp
	signedBytes, err := SignSchema([]byte(rawJSON))
	must.NoError(err)
	err = os.WriteFile(configPath, signedBytes, 0644)
	must.NoError(err)

	// 4. Tải tệp đã ký -> Phải thành công
	cfg, err := LoadConfig(configPath)
	is.NoError(err)
	is.Equal([]string{"price", "deposit"}, cfg.RequiredFields)

	// 5. Kiểm thử hàm SaveConfig
	cfg.RequiredFields = []string{"price", "deposit", "parking_fee"}
	err = SaveConfig(configPath, cfg)
	is.NoError(err)

	// Đọc trực tiếp và kiểm tra dữ liệu JSON đã lưu
	savedBytes, err := os.ReadFile(configPath)
	must.NoError(err)

	var savedConfig Config
	err = json.Unmarshal(savedBytes, &savedConfig)
	is.NoError(err)
	is.Equal([]string{"price", "deposit", "parking_fee"}, savedConfig.RequiredFields)
}
