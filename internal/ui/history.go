package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"soi-tro/internal/database"
)

func ShowHistoryAndCompareMenu() error {
	for {
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("LỊCH SỬ & SO SÁNH PHÒNG TRỌ").
					Options(
						huh.NewOption("1. So sánh song song (2-3 phòng)", "compare"),
						huh.NewOption("2. Xem danh sách lịch sử", "list"),
						huh.NewOption("3. Xóa một phòng trọ khỏi lịch sử", "delete"),
						huh.NewOption("4. Quay lại menu chính", "back"),
					).
					Value(&choice),
			),
		)

		backPressed, err := RunFormWithArrows(form)
		if err != nil {
			return err
		}
		if backPressed || choice == "back" {
			return nil
		}

		switch choice {
		case "compare":
			if err := CompareRentalsUI(); err != nil {
				fmt.Printf("❌ Lỗi đối chiếu phòng trọ: %v\n", err)
			}
		case "list":
			if err := ListRentalsUI(); err != nil {
				fmt.Printf("❌ Lỗi hiển thị danh sách phòng: %v\n", err)
			}
		case "delete":
			if err := DeleteRentalUI(); err != nil {
				fmt.Printf("❌ Lỗi xóa phòng: %v\n", err)
			}
		}
	}
}

func CompareRentalsUI() error {
	records, err := database.ListRentals()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		fmt.Println("\n⚠️  Cần tối thiểu 2 phòng trọ trong lịch sử để thực hiện đối chiếu song song!")
		return nil
	}

	var options []huh.Option[int64]
	for _, rec := range records {
		label := fmt.Sprintf("#%d - Giá: %s - ĐT: %s (%s)", rec.ID, rec.Result.Price, rec.Result.PhoneNumber, rec.CreatedAt)
		options = append(options, huh.NewOption(label, rec.ID))
	}

	var selectedIDs []int64
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int64]().
				Title("Chọn 2 đến 3 phòng trọ để so sánh").
				Options(options...).
				Value(&selectedIDs).
				Validate(func(val []int64) error {
					if len(val) < 2 || len(val) > 3 {
						return fmt.Errorf("Vui lòng chọn chính xác từ 2 đến 3 phòng")
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
		return nil
	}

	var selectedRecords []database.RentalRecord
	for _, id := range selectedIDs {
		for _, rec := range records {
			if rec.ID == id {
				selectedRecords = append(selectedRecords, rec)
				break
			}
		}
	}

	RenderComparisonTable(selectedRecords)
	return nil
}

func ListRentalsUI() error {
	records, err := database.ListRentals()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		fmt.Println("\n📭 Lịch sử trống.")
		return nil
	}

	fmt.Println("\n=========================================================================")
	fmt.Println("                    DANH SÁCH PHÒNG TRỌ ĐÃ PHÂN TÍCH                     ")
	fmt.Println("=========================================================================")
	for _, rec := range records {
		fmt.Printf("🏠 [ID #%d] Ngày: %s\n", rec.ID, rec.CreatedAt)
		fmt.Printf("   - Giá thuê: %s | Đặt cọc: %s | Tầng: %s\n", rec.Result.Price, rec.Result.Deposit, rec.Result.Floor)
		fmt.Printf("   - Liên hệ: %s | Điện: %s | Nước: %s\n", rec.Result.PhoneNumber, rec.Result.Electricity, rec.Result.Water)
		if len(rec.Result.MissingFields) > 0 {
			fmt.Printf("   - ⚠️  Thiếu thông tin: %v\n", rec.Result.MissingFields)
		}
		fmt.Println("-------------------------------------------------------------------------")
	}
	fmt.Println()
	return nil
}

func DeleteRentalUI() error {
	records, err := database.ListRentals()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		fmt.Println("\n📭 Lịch sử trống.")
		return nil
	}

	var options []huh.Option[int64]
	for _, rec := range records {
		label := fmt.Sprintf("#%d - Giá: %s - ĐT: %s (%s)", rec.ID, rec.Result.Price, rec.Result.PhoneNumber, rec.CreatedAt)
		options = append(options, huh.NewOption(label, rec.ID))
	}

	var selectedID int64
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int64]().
				Title("Chọn phòng trọ muốn xóa").
				Options(options...).
				Value(&selectedID),
		),
	)

	backPressed, err := RunFormWithArrows(form)
	if err != nil {
		return err
	}
	if backPressed {
		return nil
	}

	var confirmDelete bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Bạn có chắc chắn muốn xóa phòng trọ #%d?", selectedID)).
				Value(&confirmDelete),
		),
	)
	
	backPressed, err = RunFormWithArrows(confirmForm)
	if err != nil {
		return err
	}
	if backPressed || !confirmDelete {
		fmt.Println("Đã hủy bỏ xóa.")
		return nil
	}

	if err := database.DeleteRental(selectedID); err != nil {
		return err
	}
	fmt.Printf("✨ Đã xóa thành công phòng trọ #%d khỏi lịch sử!\n", selectedID)
	return nil
}
