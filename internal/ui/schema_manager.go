package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/genai"
	"soi-tro/internal/analyzer"
	"soi-tro/internal/gemini"
)

// ManageSchemaLoop displays the schema management UI menu and processes selections.
func ManageSchemaLoop() error {
	for {
		var manageChoice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("QUẢN LÝ CÁC TRƯỜNG THÔNG TIN (SCHEMA)").
					Options(
						huh.NewOption("1. Xem danh sách các trường hiện tại", "list"),
						huh.NewOption("2. Thêm trường thông tin mới", "add"),
						huh.NewOption("3. Xóa trường thông tin", "delete"),
						huh.NewOption("4. Bật/Tắt trạng thái bắt buộc (Required)", "toggle"),
						huh.NewOption("5. Quay lại menu chính", "back"),
					).
					Value(&manageChoice),
			),
		)
		backPressed, err := RunFormWithArrows(form)
		if err != nil {
			return err
		}
		if backPressed {
			manageChoice = "back"
		}

		if manageChoice == "back" {
			break
		}

		switch manageChoice {
		case "list":
			if err := listFields(); err != nil {
				fmt.Printf("❌ Lỗi hiển thị danh sách: %v\n", err)
			}
		case "add":
			if err := addField(); err != nil {
				fmt.Printf("❌ Lỗi khi thêm trường mới: %v\n", err)
			}
		case "delete":
			if err := deleteField(); err != nil {
				fmt.Printf("❌ Lỗi khi xóa trường: %v\n", err)
			}
		case "toggle":
			if err := toggleFieldRequired(); err != nil {
				fmt.Printf("❌ Lỗi khi đổi trạng thái: %v\n", err)
			}
		}
		fmt.Println()
	}
	return nil
}

func listFields() error {
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return err
	}
	schema, err := gemini.LoadSchema(schemaPath)
	if err != nil {
		return err
	}
	config, err := analyzer.LoadConfig(schemaPath)
	if err != nil {
		return err
	}

	fmt.Println("\n=========================================================================")
	fmt.Println("                  DANH SÁCH CÁC TRƯỜNG THÔNG TIN HIỆN TẠI                ")
	fmt.Println("=========================================================================")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Mã trường (Key)", "Tiêu đề (Title)", "Yêu cầu", "Kiểu dữ liệu", "Mô tả"})
	table.SetAutoWrapText(true)
	table.SetRowLine(true)
	table.SetColWidth(25)

	standardKeys := []string{"price", "deposit", "floor", "parking_fee", "pets_allowed", "electricity", "water", "additional_notes", "missing_fields", "sample_messages"}
	seen := make(map[string]bool)

	printRow := func(k string, prop *genai.Schema) {
		req := "Không"
		isRequired := false
		for _, r := range config.RequiredFields {
			if r == k {
				isRequired = true
				break
			}
		}
		if isRequired {
			req = "Có"
		}

		title := prop.Title
		if title == "" {
			title = "-"
		}

		table.Append([]string{k, title, req, string(prop.Type), prop.Description})
	}

	for _, k := range standardKeys {
		if prop, ok := schema.Properties[k]; ok {
			printRow(k, prop)
			seen[k] = true
		}
	}

	for k, prop := range schema.Properties {
		if !seen[k] {
			printRow(k, prop)
		}
	}

	table.Render()

	fmt.Print("\nNhấn Enter để quay lại menu quản lý...")
	var dummy string
	fmt.Scanln(&dummy)
	return nil
}

func addField() error {
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return err
	}
	schema, err := gemini.LoadSchema(schemaPath)
	if err != nil {
		return err
	}
	config, err := analyzer.LoadConfig(schemaPath)
	if err != nil {
		return err
	}

	var key string
	var title string
	var description string
	var required bool

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nhập mã trường (Chỉ dùng chữ cái viết thường, số, dấu gạch dưới)").
				Placeholder("VD: air_conditioner").
				Value(&key).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return errors.New("mã trường không được để trống")
					}
					match, _ := regexp.MatchString("^[a-z0-9_]+$", s)
					if !match {
						return errors.New("mã trường chỉ được bao gồm chữ cái viết thường, số, và dấu gạch dưới")
					}
					if _, exists := schema.Properties[s]; exists {
						return errors.New("mã trường này đã tồn tại trong schema")
					}
					return nil
				}),
			huh.NewInput().
				Title("Nhập tiêu đề hiển thị (Tiếng Việt)").
				Placeholder("VD: Điều hòa / Máy lạnh").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("tiêu đề không được để trống")
					}
					return nil
				}),
			huh.NewInput().
				Title("Nhập mô tả trường thông tin (Hướng dẫn trích xuất)").
				Placeholder("VD: Trích xuất thông tin về máy lạnh. Nếu không đề cập, ghi 'Không đề cập'.").
				Value(&description).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("mô tả không được để trống")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Có bắt buộc trích xuất trường này không (Required Checklist)?").
				Value(&required),
		),
	).Run()
	if err != nil {
		return err
	}

	key = strings.TrimSpace(key)
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	newProp := &genai.Schema{
		Type:        genai.TypeString,
		Title:       title,
		Description: description,
	}
	schema.Properties[key] = newProp

	schema.Required = append(schema.Required, key)
	if required {
		config.RequiredFields = append(config.RequiredFields, key)
	}

	if err := gemini.SaveSchema(schemaPath, schema, config.RequiredFields); err != nil {
		return fmt.Errorf("không thể ghi %s: %w", schemaPath, err)
	}

	fmt.Printf("\n✨ Đã thêm trường thông tin '%s' thành công!\n", key)
	return nil
}

func deleteField() error {
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return err
	}
	schema, err := gemini.LoadSchema(schemaPath)
	if err != nil {
		return err
	}
	config, err := analyzer.LoadConfig(schemaPath)
	if err != nil {
		return err
	}

	var options []huh.Option[string]
	for k, prop := range schema.Properties {
		if k == "missing_fields" || k == "sample_messages" || k == "additional_notes" || k == "phone_number" {
			continue
		}
		title := prop.Title
		if title == "" {
			title = k
		}
		options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", title, k), k))
	}

	if len(options) == 0 {
		fmt.Println("\nℹ️ Không có trường nào có thể xóa được.")
		return nil
	}

	var toDelete string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Chọn trường thông tin muốn xóa").
				Options(options...).
				Value(&toDelete),
		),
	)
	backPressed, err := RunFormWithArrows(form)
	if err != nil {
		return err
	}
	if backPressed {
		fmt.Println("\nĐã hủy bỏ xóa trường.")
		return nil
	}

	delete(schema.Properties, toDelete)

	var newReq []string
	for _, r := range schema.Required {
		if r != toDelete {
			newReq = append(newReq, r)
		}
	}
	schema.Required = newReq

	var newConfigReq []string
	for _, r := range config.RequiredFields {
		if r != toDelete {
			newConfigReq = append(newConfigReq, r)
		}
	}
	config.RequiredFields = newConfigReq

	if err := gemini.SaveSchema(schemaPath, schema, config.RequiredFields); err != nil {
		return err
	}

	fmt.Printf("\n✨ Đã xóa trường thông tin '%s' thành công!\n", toDelete)
	return nil
}

func toggleFieldRequired() error {
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return err
	}
	schema, err := gemini.LoadSchema(schemaPath)
	if err != nil {
		return err
	}
	config, err := analyzer.LoadConfig(schemaPath)
	if err != nil {
		return err
	}

	var options []huh.Option[string]
	var selectedFields []string

	for k, prop := range schema.Properties {
		if k == "missing_fields" || k == "sample_messages" || k == "additional_notes" || k == "phone_number" {
			continue
		}

		isRequired := false
		for _, r := range config.RequiredFields {
			if r == k {
				isRequired = true
				break
			}
		}

		title := prop.Title
		if title == "" {
			title = k
		}
		options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", title, k), k))

		if isRequired {
			selectedFields = append(selectedFields, k)
		}
	}

	if len(options) == 0 {
		fmt.Println("\nℹ️ Không có trường nào có thể bật/tắt.")
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Bật/Tắt các trường bắt buộc (Checklist Required)").
				Description("Sử dụng phím cách [Space] để Chọn/Hủy chọn, phím [Enter] để Lưu thay đổi.").
				Options(options...).
				Value(&selectedFields),
		),
	)
	backPressed, err := RunFormWithArrows(form)
	if err != nil {
		return err
	}
	if backPressed {
		fmt.Println("\nĐã hủy bỏ thay đổi.")
		return nil
	}

	// 1. Update checklist configuration
	config.RequiredFields = selectedFields

	// 2. Synchronize Gemini structural required array
	newReqLookup := make(map[string]bool)
	for _, f := range selectedFields {
		newReqLookup[f] = true
	}

	var newSchemaReq []string
	// Keep structural system fields structurally required in API contract
	systemFields := map[string]bool{
		"additional_notes": true,
		"missing_fields":   true,
		"sample_messages":  true,
		"phone_number":     true,
	}
	for f := range systemFields {
		newSchemaReq = append(newSchemaReq, f)
	}

	for k := range schema.Properties {
		if systemFields[k] {
			continue
		}
		if newReqLookup[k] {
			newSchemaReq = append(newSchemaReq, k)
		}
	}
	schema.Required = newSchemaReq

	// 3. Save updates in schema.json
	if err := gemini.SaveSchema(schemaPath, schema, config.RequiredFields); err != nil {
		return fmt.Errorf("lỗi khi lưu schema (%s): %w", schemaPath, err)
	}

	fmt.Println("\n✨ Đã lưu thay đổi trạng thái bắt buộc thành công!")
	return nil
}

// ConfigureExport lets the user set the export directory and max file size.
// It is exported so it can be called directly from the main menu.
func ConfigureExport() error {
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return err
	}
	// Load current settings to pre-populate the form.
	current, configured, err := gemini.LoadExportConfig(schemaPath)
	if err != nil {
		fmt.Printf("⚠️  Không thể tải cấu hình xuất hiện tại: %v\n", err)
	}

	// Pre-populate defaults if not yet configured.
	currentDir := current.Dir
	if !configured || currentDir == "" {
		currentDir = "."
	}
	currentMaxKB := current.MaxSizeKB
	if currentMaxKB <= 0 {
		currentMaxKB = 1024
	}

	var dirInput string
	var maxKBInput string

	dirInput = currentDir
	maxKBInput = strconv.Itoa(currentMaxKB)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Thư mục lưu file xuất kết quả").
				Description("Nhập đường dẫn thư mục. Sẽ được tạo tự động nếu chưa tồn tại.").
				Placeholder("VD: D:\\Rentals\\exports").
				Value(&dirInput).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return errors.New("đường dẫn không được để trống")
					}
					// Resolve to absolute path to validate it is a legal path.
					// This mirrors the security fix in exporter.WriteResult.
					_, err := filepath.Abs(s)
					if err != nil {
						return fmt.Errorf("đường dẫn không hợp lệ: %w", err)
					}
					return nil
				}),
			huh.NewInput().
				Title("Kích thước tối đa mỗi file (KB)").
				Description("Khi vượt quá giới hạn, file mới sẽ được tạo tự động (VD: 1024 = 1 MB).").
				Placeholder("1024").
				Value(&maxKBInput).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					n, err := strconv.Atoi(s)
					if err != nil {
						return errors.New("vui lòng nhập một số nguyên hợp lệ")
					}
					if n < 64 {
						return errors.New("kích thước tối thiểu là 64 KB")
					}
					if n > 102400 {
						return errors.New("kích thước tối đa là 102400 KB (100 MB)")
					}
					return nil
				}),
		),
	)

	backPressed, err := RunFormWithArrows(form)
	if err != nil {
		return err
	}
	if backPressed {
		fmt.Println("\nĐã hủy bỏ thay đổi cài đặt xuất.")
		return nil
	}

	dirInput = strings.TrimSpace(dirInput)
	maxKB, _ := strconv.Atoi(strings.TrimSpace(maxKBInput))

	// Resolve to absolute path before saving — consistent with exporter security fix.
	absDir, err := filepath.Abs(dirInput)
	if err != nil {
		return fmt.Errorf("đường dẫn không hợp lệ: %w", err)
	}

	// Attempt to pre-create the directory to surface permission errors immediately.
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return fmt.Errorf("không thể tạo thư mục xuất: %w", err)
	}

	if err := gemini.SaveExportConfig(schemaPath, absDir, maxKB); err != nil {
		return fmt.Errorf("không thể lưu cấu hình xuất (%s): %w", schemaPath, err)
	}

	fmt.Printf("\n✨ Đã lưu cấu hình xuất kết quả:\n")
	fmt.Printf("   Thư mục : %s\n", absDir)
	fmt.Printf("   Giới hạn: %d KB / file\n", maxKB)
	return nil
}
