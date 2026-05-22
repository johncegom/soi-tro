package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

// GlobalConfig represents the global system config format
type GlobalConfig struct {
	GeminiAPIKey string `json:"gemini_api_key"`
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

	configDir := filepath.Dir(configPath)
	// Security: Create the directory with 0o700 permissions (owner read/write/execute only)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create global config directory: %w", err)
	}

	cfg := GlobalConfig{
		GeminiAPIKey: strings.TrimSpace(key),
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

	return nil
}
