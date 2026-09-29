package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"soi-tro/internal/preferences"
)

// ManageSearchProfile edits the one local profile without changing API settings.
func ManageSearchProfile() error {
	store, err := preferences.DefaultStore()
	if err != nil {
		return err
	}
	for {
		profile, err := store.Load()
		if err != nil {
			return err
		}
		var choice string
		back, err := RunFormWithArrows(profileMenuForm(profile, &choice))
		if err != nil {
			return err
		}
		if back || choice == backChoice {
			return nil
		}
		switch choice {
		case "view":
			printProfile(*profile)
		case "edit":
			if err := editProfile(store, profile); err != nil {
				return err
			}
		case "clear":
			var confirm bool
			back, err := RunFormWithArrows(huh.NewForm(huh.NewGroup(huh.NewConfirm().
				Title("Xóa hồ sơ tìm phòng?").
				Description("←/→: chọn; Enter: xác nhận; Esc: quay lại.").
				Affirmative("Có").Negative("Không").Value(&confirm))).WithShowHelp(false))
			if err != nil {
				return err
			}
			if !back && confirm {
				if err := store.Clear(); err != nil {
					return err
				}
				fmt.Println("Đã xóa hồ sơ tìm phòng.")
			}
		}
	}
}

func profileMenuForm(profile *preferences.SearchProfile, choice *string) *huh.Form {
	options := []huh.Option[string]{huh.NewOption("Tạo / sửa hồ sơ", "edit")}
	if profile != nil {
		options = append(options, huh.NewOption("Xem hồ sơ", "view"), huh.NewOption("Xóa hồ sơ", "clear"))
	}
	options = append(options, huh.NewOption("Quay lại", backChoice))
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("HỒ SƠ TÌM PHÒNG").
		Description("Sau khi lưu, xem tin mới hoặc Lịch sử: Phù hợp = đạt; Cần kiểm tra = chưa rõ; Không phù hợp = không đạt. Tin vẫn hiển thị. ↑/↓: chọn; Enter: xác nhận; Esc: quay lại.").
		Options(options...).Value(choice))).WithShowHelp(false)
}

func editProfile(store preferences.Store, existing *preferences.SearchProfile) error {
	var p preferences.SearchProfile
	if existing != nil {
		p = *existing
	}
	rent, months, floor, unknown := formatInt64(p.MaxRentVND), formatInt(p.MaxDepositMonths), formatInt(p.MaxFloor), formatInt(p.MaxUnknown)
	form := profileEditorForm(&p, &rent, &months, &floor, &unknown)
	back, err := RunFormWithArrows(form)
	if err != nil {
		return err
	}
	if back {
		return nil
	}
	p, err = profileFromInputs(p, rent, months, floor, unknown)
	if err != nil {
		return err
	}
	if err := store.Save(p); err != nil {
		return err
	}
	fmt.Println("\nĐã lưu hồ sơ tìm phòng. Tin mới và tin trong Lịch sử sẽ được đánh giá:")
	fmt.Println("  Phù hợp: đạt các điều kiện bạn đặt.")
	fmt.Println("  Cần kiểm tra: còn thông tin thiếu hoặc chưa rõ.")
	fmt.Println("  Không phù hợp: có điều kiện không đạt hoặc quá nhiều thông tin chưa rõ.")
	fmt.Println("Tin vẫn được hiển thị cùng nhãn và lý do; không tin nào bị ẩn.")
	return nil
}

func profileFromInputs(p preferences.SearchProfile, rent, months, floor, unknown string) (preferences.SearchProfile, error) {
	var err error
	if p.MaxRentVND, err = parseOptional64(rent); err != nil {
		return p, fmt.Errorf("giá thuê: %w", err)
	}
	if p.MaxDepositMonths, err = parseOptionalInt(months); err != nil {
		return p, fmt.Errorf("số tháng cọc tối đa: %w", err)
	}
	if p.MaxDepositMonths != nil {
		// Entering a new month limit explicitly replaces the legacy cash limit.
		p.MaxMoveInCashVND = nil
	}
	if p.MaxFloor, err = parseOptionalInt(floor); err != nil {
		return p, fmt.Errorf("tầng: %w", err)
	}
	if p.MaxUnknown, err = parseOptionalInt(unknown); err != nil {
		return p, fmt.Errorf("giới hạn tiêu chí thiếu hoặc chưa rõ: %w", err)
	}
	return p, nil
}

func profileEditorForm(p *preferences.SearchProfile, rent, months, floor, unknown *string) *huh.Form {
	depositInput := huh.NewInput().Title("Cọc tối đa (tháng tiền thuê; trống = bỏ qua)").Value(months)
	if p.MaxMoveInCashVND != nil {
		depositInput.Description(fmt.Sprintf("Hồ sơ cũ: tiền thuê + cọc tối đa %s. Nhập số tháng để thay thế; để trống sẽ giữ giới hạn cũ.", profileMoney(p.MaxMoveInCashVND)))
	}
	return huh.NewForm(
		profileEditorPage(huh.NewInput().Title("Giá thuê tối đa (VND; trống = bỏ qua)").Value(rent)),
		profileEditorPage(depositInput),
		profileEditorPage(huh.NewConfirm().Title("Cần cho phép nuôi thú cưng?").Affirmative("Có").Negative("Không").Value(&p.RequirePets)),
		profileEditorPage(huh.NewConfirm().Title("Cần chỗ giữ xe?").Affirmative("Có").Negative("Không").Value(&p.RequireParking)),
		profileEditorPage(huh.NewConfirm().Title("Cần thang máy?").Affirmative("Có").Negative("Không").Value(&p.RequireElevator)),
		profileEditorPage(huh.NewInput().Title("Tầng tối đa (trống = bỏ qua)").Value(floor)),
		profileEditorPage(huh.NewInput().
			Title("Tin đăng được phép thiếu thông tin ở tối đa bao nhiêu tiêu chí?").
			Description("Thiếu hoặc không rõ đều được tính; 0 = phải rõ hết, để trống = không giới hạn.").
			Value(unknown)),
	).WithShowHelp(false)
}

func profileEditorPage(field huh.Field) *huh.Group {
	return huh.NewGroup(field).Description("Enter: tiếp hoặc lưu; Esc: quay lại (không lưu).")
}

func parseOptional64(raw string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return nil, fmt.Errorf("nhập số nguyên không âm")
	}
	return &v, nil
}

func parseOptionalInt(raw string) (*int, error) {
	v, err := parseOptional64(raw)
	if err != nil || v == nil {
		return nil, err
	}
	n := int(*v)
	if int64(n) != *v {
		return nil, fmt.Errorf("số quá lớn")
	}
	return &n, nil
}

func formatInt64(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}
func formatInt(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

func printProfile(p preferences.SearchProfile) {
	unknown := formatInt(p.MaxUnknown)
	if p.MaxUnknown == nil {
		unknown = "Không giới hạn"
	}
	floor := formatInt(p.MaxFloor)
	if p.MaxFloor == nil {
		floor = "Không đặt"
	}
	fmt.Println("\nHỒ SƠ TÌM PHÒNG")
	fmt.Printf("Giá thuê tối đa: %s\n", profileMoney(p.MaxRentVND))
	if p.MaxMoveInCashVND != nil {
		fmt.Printf("Giới hạn cũ — tiền thuê + cọc tối đa: %s (sửa hồ sơ để thay bằng số tháng cọc)\n", profileMoney(p.MaxMoveInCashVND))
	} else {
		fmt.Printf("Cọc tối đa: %s\n", profileDepositMonths(p.MaxDepositMonths))
	}
	fmt.Printf("Tầng tối đa: %s\n", floor)
	fmt.Printf("Tiêu chí thiếu hoặc chưa rõ cho phép: %s\n", unknown)
	fmt.Printf("Nuôi thú cưng: %s\n", profileRequirement(p.RequirePets))
	fmt.Printf("Chỗ giữ xe: %s\n", profileRequirement(p.RequireParking))
	fmt.Printf("Thang máy: %s\n", profileRequirement(p.RequireElevator))
}

func profileDepositMonths(v *int) string {
	if v == nil {
		return "Không đặt"
	}
	return fmt.Sprintf("%d tháng tiền thuê", *v)
}

func profileMoney(v *int64) string {
	if v == nil {
		return "Không đặt"
	}
	digits := strconv.FormatInt(*v, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "." + digits[i:]
	}
	return digits + " đồng"
}

func profileRequirement(required bool) string {
	if required {
		return "Bắt buộc"
	}
	return "Không yêu cầu"
}

// RenderProfileMatch prints all failed and unknown requirements for a listing.
func RenderProfileMatch(r preferences.Result) {
	fmt.Printf("\n🔎 Hồ sơ tìm phòng: %s\n", profileClassification(r.Classification))
	if r.MoveInCashVND != nil {
		fmt.Printf("   Tiền thuê + cọc (tham khảo, chưa tính khoản khác): %s\n", profileMoney(r.MoveInCashVND))
	}
	for _, f := range r.Findings {
		if f.Status != "pass" {
			fmt.Printf("   - %s (%s): %s\n", profileField(f.Field), profileStatus(f.Status), f.Reason)
		}
	}
}

func profileClassification(code string) string {
	switch code {
	case "Shortlist":
		return "Phù hợp"
	case "Needs checking":
		return "Cần kiểm tra"
	case "Reject":
		return "Không phù hợp"
	default:
		return "Chưa xác định"
	}
}

func profileStatus(code string) string {
	switch code {
	case "fail":
		return "Không đạt"
	case "unknown":
		return "Chưa rõ"
	default:
		return "Đạt"
	}
}

func profileField(code string) string {
	switch code {
	case "price":
		return "Giá thuê"
	case "move_in_cash":
		return "Giới hạn cũ — tiền thuê + cọc"
	case "deposit_months":
		return "Số tháng cọc"
	case "pets_allowed":
		return "Nuôi thú cưng"
	case "parking_fee":
		return "Chỗ giữ xe"
	case "elevator":
		return "Thang máy"
	case "floor":
		return "Tầng"
	case "unknown_count":
		return "Số tiêu chí thiếu hoặc chưa rõ"
	default:
		return "Tiêu chí khác"
	}
}
