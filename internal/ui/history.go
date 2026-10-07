package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"soi-tro/internal/database"
	"soi-tro/internal/preferences"

	"github.com/charmbracelet/huh"
)

const (
	deleteCancel  = "cancel"
	deleteConfirm = "delete_with_notes"
)

// ShowHistoryAndCompareMenu runs the history submenu until the user goes back.
func ShowHistoryAndCompareMenu() error {
	for {
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("LỊCH SỬ & SO SÁNH PHÒNG TRỌ").
					Options(historyMenuOptions(devModeEnabled())...).
					Value(&choice),
			),
		)

		backPressed, err := RunFormWithArrows(form)
		if err != nil {
			return err
		}
		if backPressed || choice == backChoice {
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
		case "samples":
			if err := SampleDataUI(); err != nil {
				fmt.Printf("❌ Lỗi dữ liệu mẫu: %v\n", err)
			}
		case "viewing":
			if err := ViewingTrailUI(); err != nil {
				fmt.Printf("❌ Lỗi hồ sơ xem phòng: %v\n", err)
			}
		case "delete":
			if err := DeleteRentalUI(); err != nil {
				fmt.Printf("❌ Lỗi xóa phòng: %v\n", err)
			}
		}
	}
}

// CompareRentalsUI lets the user pick 2-3 saved rentals and renders them side by side.
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
						return errors.New("vui lòng chọn chính xác từ 2 đến 3 phòng")
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

// ListRentalsUI prints every saved rental.
func ListRentalsUI() error {
	store, err := preferences.DefaultStore()
	if err != nil {
		return err
	}
	profile, err := store.Load()
	if err != nil {
		return err
	}
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
		fmt.Printf("   - 📌 Hồ sơ xem phòng: %s\n", trailSummaryFor(rec.ID))
		if len(rec.Result.MissingFields) > 0 {
			fmt.Printf("   - ⚠️  Thiếu thông tin: %v\n", rec.Result.MissingFields)
		}
		if profile != nil {
			RenderProfileMatch(preferences.Evaluate(*profile, preferences.FactsFromExtraction(rec.Result)))
		}
		fmt.Println("-------------------------------------------------------------------------")
	}
	fmt.Println()
	return nil
}

// DeleteRentalUI lets the user delete one saved rental.
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

	// Cancel is listed first so a stray Enter never deletes.
	choice := deleteCancel
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Xóa phòng trọ #%d?", selectedID)).
				Options(
					huh.NewOption("Hủy", deleteCancel),
					huh.NewOption("Xóa phòng và ghi chú xem phòng", deleteConfirm),
				).
				Value(&choice),
		),
	)

	backPressed, err = RunFormWithArrows(confirmForm)
	if err != nil {
		return err
	}
	if backPressed || choice != deleteConfirm {
		fmt.Println("Đã hủy bỏ xóa.")
		return nil
	}

	if err := database.DeleteRentalAndTrail(context.Background(), selectedID); err != nil {
		return err
	}
	fmt.Printf("✨ Đã xóa thành công phòng trọ #%d khỏi lịch sử!\n", selectedID)
	return nil
}

// devModeEnabled reports whether developer-only menu entries are shown.
func devModeEnabled() bool { return os.Getenv("SOI_TRO_DEV") == "1" }

func historyMenuOptions(dev bool) []huh.Option[string] {
	options := []huh.Option[string]{
		huh.NewOption("1. So sánh song song (2-3 phòng)", "compare"),
		huh.NewOption("2. Xem danh sách lịch sử", "list"),
		huh.NewOption("3. Hồ sơ xem phòng (checklist & quyết định)", "viewing"),
		huh.NewOption("4. Xóa một phòng trọ khỏi lịch sử", "delete"),
	}
	if dev {
		options = append(options, huh.NewOption("🧪 Dữ liệu mẫu (dev)", "samples"))
	}

	return append(options, huh.NewOption("5. Quay lại menu chính", backChoice))
}

// SampleDataUI seeds or removes fake rentals used for manual testing.
func SampleDataUI() error {
	var choice string

	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("DỮ LIỆU MẪU (chỉ để thử nghiệm, gắn nhãn "+database.SamplePrefix+")").
			Options(
				huh.NewOption("Thêm phòng mẫu", "seed"),
				huh.NewOption("Xóa tất cả phòng mẫu và hồ sơ xem phòng của chúng", "delete"),
				huh.NewOption("Quay lại", backChoice),
			).
			Value(&choice),
	))

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed {
		return err
	}

	switch choice {
	case "seed":
		n, err := database.SeedSampleRentals()
		if err != nil {
			return err
		}

		if n == 0 {
			fmt.Println("ℹ️  Phòng mẫu đã có sẵn, không thêm nữa.")
		} else {
			fmt.Printf("🧪 Đã thêm %d phòng mẫu.\n", n)
		}
	case "delete":
		n, err := database.DeleteSampleRentals()
		if err != nil {
			return err
		}

		fmt.Printf("🧹 Đã xóa %d phòng mẫu.\n", n)
	}

	return nil
}
