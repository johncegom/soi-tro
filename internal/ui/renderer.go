package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/olekukonko/tablewriter"
	"soi-tro/internal/analyzer"
	"soi-tro/internal/gemini"
)

// RenderResults displays the extraction results and configuration compliance status in the terminal.
func RenderResults(result *gemini.RentalExtractionResult, config *analyzer.Config) {
	fmt.Println("\n=========================================================================")
	fmt.Println("                 KẾT QUẢ PHÂN TÍCH TIN ĐĂNG THUÊ PHÒNG                   ")
	fmt.Println("=========================================================================")
	if result.PhoneNumber != "" && result.PhoneNumber != "Không đề cập" {
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
	schema, err := gemini.LoadSchema("schema.json")
	if err != nil {
		fmt.Printf("⚠️  Không thể tải schema.json để hiển thị tiêu đề động: %v\n", err)
	}

	type displayField struct {
		key   string
		name  string
		value string
	}
	var fields []displayField

	if schema != nil && len(schema.Properties) > 0 {
		standardOrder := []string{"price", "deposit", "floor", "parking_fee", "pets_allowed", "electricity", "water"}
		seen := make(map[string]bool)

		for _, k := range standardOrder {
			if prop, exists := schema.Properties[k]; exists {
				val := result.RawFields[k]
				title := prop.Title
				if title == "" {
					title = k
				}
				fields = append(fields, displayField{key: k, name: title, value: val})
				seen[k] = true
			}
		}

		for k, prop := range schema.Properties {
			if k == "missing_fields" || k == "sample_messages" || k == "additional_notes" || k == "phone_number" {
				continue
			}
			if !seen[k] {
				val := result.RawFields[k]
				title := prop.Title
				if title == "" {
					title = k
				}
				fields = append(fields, displayField{key: k, name: title, value: val})
			}
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
		isRequired := false
		for _, req := range config.RequiredFields {
			if req == field.key {
				isRequired = true
				break
			}
		}

		reqStr := "Không"
		if isRequired {
			reqStr = "Có (Bắt buộc)"
		}

		// Calculate compliance status
		isMissing := false
		if isRequired {
			// Check if Gemini identified it as missing
			for _, m := range result.MissingFields {
				if strings.EqualFold(m, field.key) {
					isMissing = true
					break
				}
			}
			// Double-check if the extracted value is empty, "n/a", or "không đề cập"
			valLower := strings.ToLower(strings.TrimSpace(field.value))
			if field.value == "" || valLower == "n/a" || valLower == "không đề cập" || valLower == "chưa đề cập" {
				isMissing = true
			}
		}

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
	if result.AdditionalNotes != "" && result.AdditionalNotes != "Không đề cập" {
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
	hasPhone := result.PhoneNumber != "" && result.PhoneNumber != "Không đề cập"
	if hasPhone {
		err := clipboard.WriteAll(result.PhoneNumber)
		if err != nil {
			fmt.Printf("\n⚠️  Không thể tự động sao chép số điện thoại vào clipboard: %v\n", err)
		} else {
			fmt.Printf("\n📋 [OK]: Đã tự động sao chép số điện thoại '%s' vào clipboard của bạn để tiện liên hệ!\n", result.PhoneNumber)
		}
	} else if politeMessage != "" {
		err := clipboard.WriteAll(politeMessage)
		if err != nil {
			fmt.Printf("\n⚠️  Không thể tự động sao chép tin nhắn vào clipboard: %v\n", err)
		} else {
			fmt.Println("\n📋 [OK]: Đã tự động sao chép tin nhắn phong cách 'Lịch sự' vào clipboard của bạn!")
		}
	}

	fmt.Println("=========================================================================")
}
