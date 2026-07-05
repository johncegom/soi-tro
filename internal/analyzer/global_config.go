package analyzer

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

//go:embed schema.json
var defaultSchemaBytes []byte

// GetSchemaPath returns the absolute path to ~/.config/soi-tro/schema.json
func GetSchemaPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "soi-tro", "schema.json"), nil
}

// EnsureSchemaFile checks if schema.json exists in ~/.config/soi-tro/.
// If not, it writes the default schema content.
func EnsureSchemaFile() (string, error) {
	schemaPath, err := GetSchemaPath()
	if err != nil {
		return "", err
	}

	// Check if file exists
	if _, err := os.Stat(schemaPath); err == nil {
		// Migration check: if schema exists but is not signed (e.g. legacy version), sign and lock it now.
		data, err := os.ReadFile(schemaPath)
		if err == nil {
			if VerifySchema(data) != nil {
				signedData, err := SignSchema(data)
				if err == nil {
					_ = os.WriteFile(schemaPath, signedData, 0o600)
				}
			}
		}
		return schemaPath, nil // already exists, nothing to do
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check schema file status: %w", err)
	}

	// Create config directory if not exists
	configDir := filepath.Dir(schemaPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Sign the default schema bytes before writing to ensure it's secure from the start
	signedDefaultBytes, err := SignSchema(defaultSchemaBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign default schema: %w", err)
	}

	// Write default schema bytes
	if err := os.WriteFile(schemaPath, signedDefaultBytes, 0o600); err != nil {
		return "", fmt.Errorf("failed to write default schema: %w", err)
	}

	return schemaPath, nil
}

// GlobalConfig represents the global system config format
type GlobalConfig struct {
	GeminiAPIKey string `json:"gemini_api_key"`
	Model        string `json:"model,omitempty"`
}

// GetGlobalConfigPath returns the absolute path to ~/.config/soi-tro/config.json
func GetGlobalConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "soi-tro", "config.json"), nil
}

// LoadGlobalAPIKey reads the API Key from the global config file if it exists
func LoadGlobalAPIKey() (string, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return "", err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err // Let the caller know it doesn't exist
		}
		return "", fmt.Errorf("failed to open global config file: %w", err)
	}
	defer file.Close()

	var cfg GlobalConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return "", fmt.Errorf("failed to parse global config JSON: %w", err)
	}

	return strings.TrimSpace(cfg.GeminiAPIKey), nil
}

// SaveGlobalAPIKey saves the API Key to the global config file securely
func SaveGlobalAPIKey(key string) error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Load existing configuration to preserve other fields (e.g. model)
	var existing GlobalConfig
	if fileRead, err := os.Open(configPath); err == nil {
		_ = json.NewDecoder(fileRead).Decode(&existing)
		fileRead.Close()
	}

	configDir := filepath.Dir(configPath)
	// Security: Create the directory with 0o700 permissions (owner read/write/execute only)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create global config directory: %w", err)
	}

	cfg := GlobalConfig{
		GeminiAPIKey: strings.TrimSpace(key),
		Model:        existing.Model,
	}

	// Security: Write the file with 0o600 permissions (owner read/write only) to protect the API key
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open global config file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&cfg); err != nil {
		return fmt.Errorf("failed to encode global config JSON: %w", err)
	}

	return nil
}

// GetGlobalModel returns the configured model from the global config file, or the default if not set/invalid
func GetGlobalModel() string {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return "gemini-3.1-flash-lite"
	}
	file, err := os.Open(configPath)
	if err != nil {
		return "gemini-3.1-flash-lite"
	}
	defer file.Close()

	var cfg GlobalConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return "gemini-3.1-flash-lite"
	}
	if cfg.Model == "" {
		return "gemini-3.1-flash-lite"
	}
	return cfg.Model
}

// SaveGlobalModel writes the selected model into the global config, preserving other fields.
func SaveGlobalModel(model string) error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	var cfg GlobalConfig
	file, err := os.Open(configPath)
	if err == nil {
		_ = json.NewDecoder(file).Decode(&cfg)
		file.Close()
	}

	cfg.Model = strings.TrimSpace(model)

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create global config directory: %w", err)
	}

	fileWrite, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open global config file for writing: %w", err)
	}
	defer fileWrite.Close()

	encoder := json.NewEncoder(fileWrite)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&cfg); err != nil {
		return fmt.Errorf("failed to encode global config JSON: %w", err)
	}

	return nil
}

// PromptAndSaveModel prompts the user to select their desired Gemini model and saves it.
func PromptAndSaveModel() error {
	var selectedModel string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Chọn mô hình Gemini để sử dụng (Mặc định: gemini-3.1-flash-lite)").
				Options(
					huh.NewOption("Gemini 3.5 Flash (Khuyên dùng)", "gemini-3.5-flash"),
					huh.NewOption("Gemini 3.1 Pro (Preview)", "gemini-3.1-pro-preview"),
					huh.NewOption("Gemini 3.1 Flash-Lite (Cực nhanh & siêu rẻ)", "gemini-3.1-flash-lite"),
					huh.NewOption("Gemini 2.5 Flash", "gemini-2.5-flash"),
					huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
				).
				Value(&selectedModel),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("hủy bỏ chọn mô hình: %w", err)
	}

	if selectedModel == "" {
		selectedModel = "gemini-3.1-flash-lite"
	}

	if err := SaveGlobalModel(selectedModel); err != nil {
		return fmt.Errorf("không thể lưu mô hình: %w", err)
	}

	fmt.Printf("\n✨ Đã thiết lập mô hình sử dụng thành công: %s\n\n", selectedModel)
	return nil
}

// EnsureGlobalAPIKey ensures GEMINI_API_KEY is loaded into the environment
func EnsureGlobalAPIKey() error {
	// 1. Check if the key is already in the environment
	if os.Getenv("GEMINI_API_KEY") != "" {
		return nil
	}

	// 2. Try to load from the global config file
	key, err := LoadGlobalAPIKey()
	if err == nil && key != "" {
		os.Setenv("GEMINI_API_KEY", key)
		return nil
	}

	// 3. Prompt the user using Huh if not found anywhere
	fmt.Println("\n=========================================================================")
	fmt.Println("🔑                     THIẾT LẬP GEMINI API KEY                          ")
	fmt.Println("=========================================================================")
	fmt.Println("Chưa tìm thấy GEMINI_API_KEY trong tệp .env hoặc cấu hình hệ thống.")
	fmt.Println("Bạn có thể nhận khóa API miễn phí tại: https://aistudio.google.com/")
	fmt.Println("=========================================================================")

	var apiKeyInput string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Vui lòng dán Gemini API Key của bạn").
				Placeholder("VD: AIzaSy...").
				EchoMode(huh.EchoModePassword). // Hide the key as it is typed
				Value(&apiKeyInput).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return errors.New("API Key không được để trống")
					}
					if !strings.HasPrefix(s, "AIzaSy") {
						return errors.New("API Key Google Gemini thường bắt đầu bằng 'AIzaSy'")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("hủy bỏ thiết lập API Key: %w", err)
	}

	apiKeyInput = strings.TrimSpace(apiKeyInput)

	// Save the key globally
	if err := SaveGlobalAPIKey(apiKeyInput); err != nil {
		return fmt.Errorf("không thể lưu API Key toàn cục: %w", err)
	}

	// Set in environment for current process execution
	os.Setenv("GEMINI_API_KEY", apiKeyInput)

	fmt.Println("\n✨ Đã lưu cấu hình API Key thành công vào: ~\\.config\\soi-tro\\config.json")
	fmt.Println("Ứng dụng sẽ tự động sử dụng khóa này từ nay về sau!")
	fmt.Println("=========================================================================")
	fmt.Println()

	// Prompt the user for model choice immediately after API Key config
	if err := PromptAndSaveModel(); err != nil {
		fmt.Printf("⚠️ Không thể thiết lập mô hình, sẽ sử dụng mặc định (gemini-3.1-flash-lite): %v\n", err)
	}

	return nil
}
