// Command soi-tro is an interactive CLI that analyzes Vietnamese rental posts with Gemini.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"soi-tro/internal/analyzer"
	"soi-tro/internal/database"
	"soi-tro/internal/exporter"
	"soi-tro/internal/gemini"
	"soi-tro/internal/logger"
	"soi-tro/internal/ui"

	"github.com/charmbracelet/huh"
)

func main() {
	// Initialize logger
	logCfg := logger.LoadFromEnv()
	if err := logger.Init(logCfg); err != nil {
		log.Fatalf("Không thể khởi tạo logger: %v", err)
	}
	logger.Info("application starting", "operation", "application.start", "version", "1.0.0")

	// Create context with request ID for this session
	ctx := logger.NewContext(context.Background())
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

// run initializes local state and serves the main menu until the user exits.
func run(ctx context.Context) error {
	sessionLog := logger.FromContext(ctx)

	// Load environment variables from the secure .env file
	_ = analyzer.LoadEnv(".env")
	sessionLog.Info("environment load completed", "operation", "config.load_env")

	// Initialize the history database
	if err := database.InitDB(); err != nil {
		return fmt.Errorf("❌ Lỗi khởi tạo cơ sở dữ liệu lịch sử: %w", err)
	}

	// Ensure GEMINI_API_KEY is available (check env, load from global config, or prompt)
	if err := analyzer.EnsureGlobalAPIKey(); err != nil {
		sessionLog.Error("API key configuration failed", "operation", "config.api_key", "error", "configure_api_key")
		return fmt.Errorf("❌ Lỗi cấu hình API Key: %w", err)
	}
	sessionLog.Info("API key configuration completed", "operation", "config.api_key")

	// Ensure schema.json is initialized in ~/.config/soi-tro/
	schemaPath, err := analyzer.EnsureSchemaFile()
	if err != nil {
		sessionLog.Error("schema initialization failed", "operation", "schema.ensure", "error", "initialize_schema")
		return fmt.Errorf("❌ Lỗi cấu hình tệp cấu hình (schema.json): %w", err)
	}
	sessionLog.Info("schema initialization completed", "operation", "schema.ensure")

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
			sessionLog.Error("main menu selection failed", "operation", "ui.main_menu", "error", "select_action")
			return fmt.Errorf("❌ Lỗi chọn chức năng chính: %w", err)
		}
		if backPressed {
			mainChoice = "exit"
		}

		switch dispatch(mainChoice) {
		case actionExit:
			fmt.Println("\nCảm ơn bạn đã sử dụng Soi Trọ! Tạm biệt.")
			return nil
		case actionHistory:
			if err := ui.ShowHistoryAndCompareMenu(); err != nil {
				sessionLog.Error("history view failed", "operation", "ui.history", "error", "show_history")
				fmt.Printf("❌ Lỗi hiển thị lịch sử: %v\n", err)
			}
		case actionManage:
			if err := ui.ManageSchemaLoop(); err != nil {
				sessionLog.Error("schema management failed", "operation", "ui.schema", "error", "manage_schema")
				fmt.Printf("❌ Lỗi quản lý schema: %v\n", err)
			}
		case actionExport:
			if err := ui.ConfigureExport(); err != nil {
				sessionLog.Error("export configuration failed", "operation", "ui.export_config", "error", "configure_export")
				fmt.Printf("❌ Lỗi cài đặt xuất kết quả: %v\n", err)
			}
		case actionModel:
			if err := analyzer.PromptAndSaveModel(); err != nil {
				sessionLog.Error("model configuration failed", "operation", "ui.model_config", "error", "configure_model")
				fmt.Printf("❌ Lỗi cấu hình mô hình: %v\n", err)
			}
		case actionAnalyze:
			if err := runAnalyze(ctx, schemaPath); err != nil {
				return err
			}
		}
	}
}

// runAnalyze drives the "analyze new listing" flow until the user goes back to the menu.
//
//nolint:gocyclo // moved verbatim out of main; the retry/export loops share labels, splitting them is a separate change.
func runAnalyze(ctx context.Context, schemaPath string) error {
	sessionLog := logger.FromContext(ctx)

	// 1. Load configuration from schema.json
	cfg, err := analyzer.LoadConfig(schemaPath)
	if err != nil {
		sessionLog.Error("schema configuration load failed", "operation", "schema.load_config", "error", "load_config")
		return fmt.Errorf("❌ Lỗi tải tệp cấu hình (%s): %w", schemaPath, err)
	}
	sessionLog.Info("schema configuration load completed", "operation", "schema.load_config")

	// Loop to manage the input mode & retries
analyzeLoop:
	for {
		// 2. Display interactive CLI Forms to gather input (text listing or image file path)
		inputRes, err := ui.GetUserInput()
		if err != nil {
			if errors.Is(err, ui.ErrGoBack) {
				break analyzeLoop // back to main menu
			}
			sessionLog.Error("listing input failed", "operation", "ui.listing_input", "error", "read_input")
			log.Printf("❌ Lỗi nhận dữ liệu đầu vào: %v", err)
			break analyzeLoop
		}

		// Inner loop to retry analyzing the same inputRes
		for {
			analysisCtx := logger.NewContext(context.Background())

			// 3. Initialize Gemini API Client
			client, err := gemini.NewClient(analysisCtx)
			if err != nil {
				fmt.Println("\n❌ LỖI KHỞI TẠO CLIENT GEMINI:")
				fmt.Println("   Hãy đảm bảo bạn đã thiết lập biến môi trường GEMINI_API_KEY.")
				fmt.Println("   Bạn có thể điền thông tin vào tệp bảo mật .env:")
				fmt.Println("     GEMINI_API_KEY=\"KHÓA_TỪ_GOOGLE_AI_STUDIO\"")
				fmt.Println("   Hoặc chạy qua PowerShell:")
				fmt.Println("     $env:GEMINI_API_KEY=\"KHÓA_TỪ_GOOGLE_AI_STUDIO\"")
				return fmt.Errorf("khởi tạo client Gemini: %w", err)
			}

			var imageBytes []byte
			var mimeType string

			// Handle multimodal input if an image file path is chosen
			if inputRes.Type == ui.InputTypeImage {
				fmt.Printf("\n📂 Đang tải file hình ảnh: %s...\n", inputRes.ImagePath)
				imageBytes, err = os.ReadFile(inputRes.ImagePath)
				if err != nil {
					logger.FromContext(analysisCtx).Error("image read failed", "operation", "input.read_image", "error", "read_image")
					fmt.Printf("❌ Lỗi khi đọc file hình ảnh: %v\n", err)
					switch onAnalysisError(ui.PromptErrorRetry(inputRes.Type)) {
					case stepRetry:
						continue
					case stepChangeInput:
						continue analyzeLoop // prompt for input again
					default:
						break analyzeLoop // back to main menu
					}
				}

				mimeType = mimeTypeFor(inputRes.ImagePath)
			}

			// 4. Send the payload to Gemini and run extraction
			modelName := analyzer.GetGlobalModel()
			fmt.Printf("\n🤖 Đang phân tích thông tin bằng %s... Vui lòng đợi.\n", modelName)
			result, err := client.ExtractRentalInfo(analysisCtx, inputRes.Text, imageBytes, mimeType, cfg.RequiredFields)
			if err != nil {
				fmt.Printf("\n❌ Lỗi phân tích tin đăng qua Gemini API: %v\n", err)
				switch onAnalysisError(ui.PromptErrorRetry(inputRes.Type)) {
				case stepRetry:
					continue
				case stepChangeInput:
					continue analyzeLoop // prompt for input again
				default:
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
				logger.FromContext(analysisCtx).Error("export configuration load failed", "operation", "schema.load_export_config", "error", "load_export_config")
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
				} else {
					logger.FromContext(analysisCtx).Error("export title schema load failed", "operation", "schema.load_export_titles", "error", "load_schema")
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
				switch onAnalysisSuccess(nextChoice) {
				case stepExport:
					if writeErr := exporter.WriteResult(writeCfg, result, titleMap); writeErr != nil {
						fmt.Printf("⚠️  Không thể xuất kết quả: %v\n", writeErr)
					} else if filePath, pathErr := exporter.ActiveFilePath(writeCfg); pathErr == nil {
						fmt.Printf("💾 Đã lưu kết quả vào: %s\n", filePath)
						exportDone = true
					}
					// Stay in the loop so the user can pick another action.
				case stepNewInput:
					continue analyzeLoop // prompt for input again
				default: // stepBackToMenu
					break analyzeLoop
				}
			}
		}
	}
	return nil
}
