package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"soi-tro/internal/analyzer"
	"soi-tro/internal/database"
	"soi-tro/internal/exporter"
	"soi-tro/internal/gemini"
	"soi-tro/internal/ui"

	"github.com/charmbracelet/huh"
)

func main() {
	// Load environment variables from the secure .env file
	_ = analyzer.LoadEnv(".env")

	// Initialize the history database
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Lỗi khởi tạo cơ sở dữ liệu lịch sử: %v", err)
	}

	// Ensure GEMINI_API_KEY is available (check env, load from global config, or prompt)
	if err := analyzer.EnsureGlobalAPIKey(); err != nil {
		log.Fatalf("❌ Lỗi cấu hình API Key: %v", err)
	}

	// Ensure schema.json is initialized in ~/.config/soi-tro/
	schemaPath, err := analyzer.EnsureSchemaFile()
	if err != nil {
		log.Fatalf("❌ Lỗi cấu hình tệp cấu hình (schema.json): %v", err)
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
						huh.NewOption("2. Xem lịch sử & So sánh phòng trọ", "history"),
						huh.NewOption("3. Quản lý các trường thông tin (Schema)", "manage"),
						huh.NewOption("4. Cài đặt xuất kết quả (Export)", "export"),
						huh.NewOption("5. Cấu hình mô hình Gemini (Model)", "model"),
						huh.NewOption("6. Thoát", "exit"),
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

		if mainChoice == "history" {
			if err := ui.ShowHistoryAndCompareMenu(); err != nil {
				fmt.Printf("❌ Lỗi hiển thị lịch sử: %v\n", err)
			}
			continue
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

		if mainChoice == "model" {
			if err := analyzer.PromptAndSaveModel(); err != nil {
				fmt.Printf("❌ Lỗi cấu hình mô hình: %v\n", err)
			}
			continue
		}
		// 1. Load configuration from schema.json
		cfg, err := analyzer.LoadConfig(schemaPath)
		if err != nil {
			log.Fatalf("❌ Lỗi tải tệp cấu hình (%s): %v", schemaPath, err)
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
				modelName := analyzer.GetGlobalModel()
				fmt.Printf("\n🤖 Đang phân tích thông tin bằng %s... Vui lòng đợi.\n", modelName)
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
				if dbID, dbErr := database.SaveRental(result); dbErr == nil {
					fmt.Printf("💾 Đã lưu kết quả phân tích vào lịch sử (Mã số: #%d)\n", dbID)
				}
				fmt.Println()

				// 6. Load export config once to determine if export option should be shown.
				exportCfg, exportConfigured, exportErr := gemini.LoadExportConfig(schemaPath)
				if exportErr != nil {
					fmt.Printf("⚠️  Không thể đọc cấu hình xuất: %v\n", exportErr)
					exportConfigured = false
				}

				// Build title map once for reuse across possible re-exports.
				titleMap := map[string]string{}
				if exportConfigured {
					if schema, schemaErr := gemini.LoadSchema(schemaPath); schemaErr == nil {
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
