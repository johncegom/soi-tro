package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"soi-tro/internal/analyzer"
	"soi-tro/internal/exporter"
	"soi-tro/internal/gemini"
	"soi-tro/internal/ui"

	"github.com/charmbracelet/huh"
)

// loadEnv reads a .env file and sets environment variables in a secure manner.
func loadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// If .env is missing, we gracefully proceed and rely on standard shell environment variables.
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove wrapping quotes if present
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		os.Setenv(key, val)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Warning: error scanning environment file %s: %v", filename, err)
	}
}

func main() {
	// Load environment variables from the secure .env file
	loadEnv(".env")

	// Ensure GEMINI_API_KEY is available (check env, load from global config, or prompt)
	if err := analyzer.EnsureGlobalAPIKey(); err != nil {
		log.Fatalf("❌ Lỗi cấu hình API Key: %v", err)
	}

	for {
		fmt.Println("=========================================================================")
		fmt.Println("🚀                  SOI TRỌ - TRÍCH XUẤT TIN THUÊ PHÒNG                  ")
		fmt.Println("=========================================================================")

		var mainChoice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("BẢNG ĐIỀU KHIỂN - SOI TRỌ").
					Options(
						huh.NewOption("1. Phân tích tin đăng mới", "analyze"),
						huh.NewOption("2. Quản lý các trường thông tin (Schema)", "manage"),
						huh.NewOption("3. Cài đặt xuất kết quả (Export)", "export"),
						huh.NewOption("4. Thoát", "exit"),
					).
					Value(&mainChoice),
			),
		)
		backPressed, err := ui.RunFormWithArrows(form)
		if err != nil {
			log.Fatalf("❌ Lỗi chọn chức năng chính: %v", err)
		}
		if backPressed {
			mainChoice = "exit"
		}

		if mainChoice == "exit" {
			fmt.Println("\nCảm ơn bạn đã sử dụng Soi Trọ! Tạm biệt.")
			break
		}

		if mainChoice == "manage" {
			if err := ui.ManageSchemaLoop(); err != nil {
				fmt.Printf("❌ Lỗi quản lý schema: %v\n", err)
			}
			continue
		}

		if mainChoice == "export" {
			if err := ui.ConfigureExport(); err != nil {
				fmt.Printf("❌ Lỗi cài đặt xuất kết quả: %v\n", err)
			}
			continue
		}
		// 1. Load configuration from schema.json
		cfg, err := analyzer.LoadConfig("schema.json")
		if err != nil {
			log.Fatalf("❌ Lỗi tải tệp cấu hình (schema.json): %v", err)
		}

		// Loop to manage the input mode & retries
	analyzeLoop:
		for {
			// 2. Display interactive CLI Forms to gather input (text listing or image file path)
			inputRes, err := ui.GetUserInput()
			if err != nil {
				if errors.Is(err, ui.ErrGoBack) {
					break analyzeLoop // back to main menu
				}
				log.Printf("❌ Lỗi nhận dữ liệu đầu vào: %v", err)
				break analyzeLoop
			}

			// Inner loop to retry analyzing the same inputRes
			for {
				ctx := context.Background()

				// 3. Initialize Gemini API Client
				client, err := gemini.NewClient(ctx)
				if err != nil {
					fmt.Println("\n❌ LỖI KHỞI TẠO CLIENT GEMINI:")
					fmt.Println("   Hãy đảm bảo bạn đã thiết lập biến môi trường GEMINI_API_KEY.")
					fmt.Println("   Bạn có thể điền thông tin vào tệp bảo mật .env:")
					fmt.Println("     GEMINI_API_KEY=\"AIzaSy...\"")
					fmt.Println("   Hoặc chạy qua PowerShell:")
					fmt.Println("     $env:GEMINI_API_KEY=\"AIzaSy...\"")
					os.Exit(1)
				}

				var imageBytes []byte
				var mimeType string

				// Handle multimodal input if an image file path is chosen
				if inputRes.Type == ui.InputTypeImage {
					fmt.Printf("\n📂 Đang tải file hình ảnh: %s...\n", inputRes.ImagePath)
					imageBytes, err = os.ReadFile(inputRes.ImagePath)
					if err != nil {
						fmt.Printf("❌ Lỗi khi đọc file hình ảnh: %v\n", err)
						retryChoice := ui.PromptErrorRetry(inputRes.Type)
						if retryChoice == "retry" {
							continue
						} else if retryChoice == "change" {
							break // break inner loop to prompt for input again
						} else {
							break analyzeLoop // back to main menu
						}
					}

					lowerPath := strings.ToLower(inputRes.ImagePath)
					if strings.HasSuffix(lowerPath, ".png") {
						mimeType = "image/png"
					} else if strings.HasSuffix(lowerPath, ".webp") {
						mimeType = "image/webp"
					} else {
						mimeType = "image/jpeg" // .jpg hoặc .jpeg
					}
				}

				// 4. Send the payload to Gemini and run extraction
				fmt.Println("\n🤖 Đang phân tích tin thông tin bằng Gemini 2.5 Flash... Vui lòng đợi.")
				result, err := client.ExtractRentalInfo(ctx, inputRes.Text, imageBytes, mimeType, cfg.RequiredFields)
				if err != nil {
					fmt.Printf("\n❌ Lỗi phân tích tin đăng qua Gemini API: %v\n", err)
					retryChoice := ui.PromptErrorRetry(inputRes.Type)
					if retryChoice == "retry" {
						continue
					} else if retryChoice == "change" {
						break // break inner loop to prompt for input again
					} else {
						break analyzeLoop // back to main menu
					}
				}

				// 5. Format and print the final compliance table & handle clipboard copying
				ui.RenderResults(result, cfg)
				fmt.Println()

				// 6. Load export config once to determine if export option should be shown.
				exportCfg, exportConfigured, exportErr := gemini.LoadExportConfig("schema.json")
				if exportErr != nil {
					fmt.Printf("⚠️  Không thể đọc cấu hình xuất: %v\n", exportErr)
					exportConfigured = false
				}

				// Build title map once for reuse across possible re-exports.
				titleMap := map[string]string{}
				if exportConfigured {
					if schema, schemaErr := gemini.LoadSchema("schema.json"); schemaErr == nil {
						for k, prop := range schema.Properties {
							if prop.Title != "" {
								titleMap[k] = prop.Title
							}
						}
					}
				}

				writeCfg := exporter.Config{
					Dir:       exportCfg.Dir,
					MaxSizeKB: exportCfg.MaxSizeKB,
				}

				// Post-result action loop — stays here until the user picks new/back.
				exportDone := false
				for {
					nextChoice := ui.PromptAfterSuccess(exportConfigured, exportDone)
					switch nextChoice {
					case "export":
						if writeErr := exporter.WriteResult(writeCfg, result, titleMap); writeErr != nil {
							fmt.Printf("⚠️  Không thể xuất kết quả: %v\n", writeErr)
						} else if filePath, pathErr := exporter.ActiveFilePath(writeCfg); pathErr == nil {
							fmt.Printf("💾 Đã lưu kết quả vào: %s\n", filePath)
							exportDone = true
						}
						// Stay in the loop so the user can pick another action.
					case "new":
						goto nextInput
					default: // "back"
						break analyzeLoop
					}
				}
			nextInput:
			}
		}
	}
}
