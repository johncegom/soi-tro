package ui

import (
	"fmt"
	"os"
	"slices"
	"soi-tro/internal/analyzer"
	"soi-tro/internal/database"
	"soi-tro/internal/gemini"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/genai"
)

// notMentioned is the sentinel Gemini returns for absent fields.
const notMentioned = "Không đề cập"

// RenderResults displays the extraction results and configuration compliance status in the terminal.
//
//nolint:gocyclo,funlen // one branch per rental field; a table-driven rewrite is cosmetic until a field has a bug.
func RenderResults(result *gemini.RentalExtractionResult, config *analyzer.Config) {
	fmt.Println("\n=========================================================================")
	fmt.Println("                 KẾT QUẢ PHÂN TÍCH TIN ĐĂNG THUÊ PHÒNG                   ")
	fmt.Println("=========================================================================")
	if result.PhoneNumber != "" && result.PhoneNumber != notMentioned {
		fmt.Printf("📞 LIÊN HỆ CHỦ NHÀ: %s\n", result.PhoneNumber)
	} else {
		fmt.Println("📞 LIÊN HỆ CHỦ NHÀ: Không đề cập")
	}
	fmt.Println("=========================================================================")

	// Setup table writer
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Thuộc Tính", "Giá Trị Trích Xuất", "Yêu Cầu", "Trạng Thái"})
	table.SetAutoWrapText(true)
	table.SetRowLine(true)
	table.SetColWidth(35)

	// Load schema to get properties dynamically
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		fmt.Printf("⚠️  Không thể lấy đường dẫn schema: %v\n", err)
		return
	}
	schema, err := gemini.LoadSchema(schemaPath)
	if err != nil {
		fmt.Printf("⚠️  Không thể tải %s để hiển thị tiêu đề động: %v\n", schemaPath, err)
	}

	type displayField struct {
		key   string
		name  string
		value string
	}
	var fields []displayField

	if schema != nil && len(schema.Properties) > 0 {
		for _, k := range orderedKeys(schema.Properties, standardFieldKeys, nonDisplayKeys...) {
			title := schema.Properties[k].Title
			if title == "" {
				title = k
			}

			fields = append(fields, displayField{key: k, name: title, value: result.RawFields[k]})
		}
	} else {
		fields = []displayField{
			{"price", "Giá thuê", result.Price},
			{"deposit", "Tiền đặt cọc", result.Deposit},
			{"floor", "Số tầng / Lầu", result.Floor},
			{"parking_fee", "Phí giữ xe", result.ParkingFee},
			{"pets_allowed", "Cho phép nuôi thú cưng", result.PetsAllowed},
			{"electricity", "Tiền điện", result.Electricity},
			{"water", "Tiền nước", result.Water},
		}
	}

	for _, field := range fields {
		isRequired := slices.Contains(config.RequiredFields, field.key)

		reqStr := "Không"
		if isRequired {
			reqStr = "Có (Bắt buộc)"
		}

		// Calculate compliance status
		isMissing := fieldMissing(isRequired, result.MissingFields, field.key, field.value)

		status := "[ OK ]"
		if isMissing {
			status = "[ THIẾU ]"
		}

		row := []string{field.name, field.value, reqStr, status}
		table.Append(row)
	}

	table.Render()

	// Print Additional Notes
	fmt.Println("\n📝 GHI CHÚ THÊM:")
	if result.AdditionalNotes != "" && result.AdditionalNotes != notMentioned {
		fmt.Printf("  - %s\n", result.AdditionalNotes)
	} else {
		fmt.Println("  - Không có ghi chú thêm.")
	}

	// Print Missing Fields checklist analysis summary
	fmt.Println("\n🔍 DANH SÁCH CÁC TRƯỜNG THIẾU CẦN HỎI THÊM:")
	if len(result.MissingFields) > 0 {
		for _, m := range result.MissingFields {
			displayLabel := m
			if schema != nil {
				if prop, exists := schema.Properties[strings.ToLower(m)]; exists && prop.Title != "" {
					displayLabel = fmt.Sprintf("%s (%s)", prop.Title, m)
				}
			}
			fmt.Printf("  - %s\n", displayLabel)
		}
	} else {
		fmt.Println("  - Đầy đủ tất cả các trường bắt buộc của cấu hình checklist!")
	}

	// Print Sample Messages
	fmt.Println("\n💬 DỰ THẢO TIN NHẮN HỎI THÊM:")
	var politeMessage string
	for _, msg := range result.SampleMessages {
		fmt.Printf("\n👉 [%s]:\n\"%s\"\n", msg.Style, msg.Content)
		if msg.Style == "Lịch sự" {
			politeMessage = msg.Content
		}
	}

	// Copy phone number or polite message to system clipboard automatically
	hasPhone := result.PhoneNumber != "" && result.PhoneNumber != notMentioned
	switch {
	case hasPhone:
		if err := clipboard.WriteAll(result.PhoneNumber); err != nil {
			fmt.Printf("\n⚠️  Không thể tự động sao chép số điện thoại vào clipboard: %v\n", err)
		} else {
			fmt.Printf("\n📋 [OK]: Đã tự động sao chép số điện thoại '%s' vào clipboard của bạn để tiện liên hệ!\n", result.PhoneNumber)
		}
	case politeMessage != "":
		if err := clipboard.WriteAll(politeMessage); err != nil {
			fmt.Printf("\n⚠️  Không thể tự động sao chép tin nhắn vào clipboard: %v\n", err)
		} else {
			fmt.Println("\n📋 [OK]: Đã tự động sao chép tin nhắn phong cách 'Lịch sự' vào clipboard của bạn!")
		}
	}

	fmt.Println("=========================================================================")
}

// RenderComparisonTable displays 2 or 3 rental records side by side in a formatted table
//
//nolint:gocyclo // same shape as RenderResults, one branch per compared field.
func RenderComparisonTable(records []database.RentalRecord) {
	fmt.Println("\n=========================================================================")
	fmt.Println("                       SO SÁNH PHÒNG TRỌ SONG SONG                       ")
	fmt.Println("=========================================================================")

	table := tablewriter.NewWriter(os.Stdout)

	headers := []string{"Tiêu chí"}
	for i, rec := range records {
		headers = append(headers, fmt.Sprintf("Phòng %d (ID: %d)", i+1, rec.ID))
	}
	table.SetHeader(headers)
	table.SetAutoWrapText(true)
	table.SetRowLine(true)
	table.SetColWidth(30)

	schemaPath, err := analyzer.GetSchemaPath()
	var schema *genai.Schema
	if err == nil {
		schema, _ = gemini.LoadSchema(schemaPath)
	}

	type compareField struct {
		name string
		key  string
	}
	var fields []compareField

	if schema != nil && len(schema.Properties) > 0 {
		for _, k := range orderedKeys(schema.Properties, standardFieldKeys, nonDisplayKeys...) {
			title := schema.Properties[k].Title
			if title == "" {
				title = k
			}

			fields = append(fields, compareField{name: title, key: k})
		}
	} else {
		fields = []compareField{
			{"Giá thuê", "price"},
			{"Tiền đặt cọc", "deposit"},
			{"Số tầng / Lầu", "floor"},
			{"Phí giữ xe", "parking_fee"},
			{"Cho nuôi thú cưng", "pets_allowed"},
			{"Tiền điện", "electricity"},
			{"Tiền nước", "water"},
		}
	}

	rowDate := []string{"Ngày phân tích"}
	for _, rec := range records {
		rowDate = append(rowDate, rec.CreatedAt)
	}
	table.Append(rowDate)

	rowPhone := []string{"Liên hệ chủ nhà"}
	for _, rec := range records {
		phone := rec.Result.PhoneNumber
		if phone == "" {
			phone = notMentioned
		}
		rowPhone = append(rowPhone, phone)
	}
	table.Append(rowPhone)

	for _, field := range fields {
		row := []string{field.name}
		for _, rec := range records {
			val := rec.Result.RawFields[field.key]
			if val == "" {
				val = notMentioned
			}
			row = append(row, val)
		}
		table.Append(row)
	}

	rowMissing := []string{"Thiếu thông tin"}
	for _, rec := range records {
		var missingLabels []string
		for _, m := range rec.Result.MissingFields {
			lbl := m
			if schema != nil {
				if prop, exists := schema.Properties[strings.ToLower(m)]; exists && prop.Title != "" {
					lbl = prop.Title
				}
			}
			missingLabels = append(missingLabels, lbl)
		}
		if len(missingLabels) == 0 {
			rowMissing = append(rowMissing, "Đầy đủ")
		} else {
			rowMissing = append(rowMissing, strings.Join(missingLabels, ", "))
		}
	}
	table.Append(rowMissing)

	rowNotes := []string{"Ghi chú thêm"}
	for _, rec := range records {
		notes := rec.Result.AdditionalNotes
		if notes == "" || notes == notMentioned {
			notes = "-"
		}
		rowNotes = append(rowNotes, notes)
	}
	table.Append(rowNotes)

	table.Render()
	fmt.Println("=========================================================================")
}
