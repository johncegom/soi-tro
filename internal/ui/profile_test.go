package ui

import (
	"io"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"soi-tro/internal/preferences"
)

func TestProfileEditorFirstInputTitleVisible(t *testing.T) {
	rent := "5000000"
	var cash, floor, unknown string
	form := profileEditorForm(&preferences.SearchProfile{}, &rent, &cash, &floor, &unknown)
	form.Init()
	form.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if view := form.View(); !strings.Contains(view, "Giá thuê tối đa") {
		t.Fatalf("first input title missing: %q", view)
	}
	if view := form.View(); !strings.Contains(view, "Esc: quay lại") {
		t.Fatalf("back navigation hint missing: %q", view)
	}
	form.NextGroup()
	if view := form.View(); !strings.Contains(view, "Cọc tối đa (tháng tiền thuê") {
		t.Fatalf("next input title missing: %q", view)
	}
}

func TestProfileMenuExplainsWhatHappensBeforeSetup(t *testing.T) {
	for _, profile := range []*preferences.SearchProfile{nil, {RequirePets: true}} {
		var choice string
		form := profileMenuForm(profile, &choice)
		form.Init()
		form.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
		view := strings.Join(strings.Fields(strings.ReplaceAll(form.View(), "┃", "")), " ")
		if !strings.Contains(view, "Sau khi lưu") || !strings.Contains(view, "Lịch sử") ||
			!strings.Contains(view, "Phù hợp") || !strings.Contains(view, "Cần kiểm tra") ||
			!strings.Contains(view, "Không phù hợp") ||
			!strings.Contains(view, "Tin vẫn hiển thị") || !strings.Contains(view, "Tạo / sửa hồ sơ") {
			t.Fatalf("profile menu lacks an explanation at compact size: %q", view)
		}
	}
}

func captureProfileOutput(t *testing.T, show func()) string {
	t.Helper()
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = writeEnd
	defer func() { os.Stdout = old; _ = readEnd.Close(); _ = writeEnd.Close() }()
	show()
	os.Stdout = old
	if err := writeEnd.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(readEnd)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestProfileViewUsesVietnameseValues(t *testing.T) {
	rent, months, floor := int64(5_500_000), 1, 2
	output := captureProfileOutput(t, func() {
		printProfile(preferences.SearchProfile{MaxRentVND: &rent, MaxDepositMonths: &months, MaxFloor: &floor, RequireParking: true, RequireElevator: true})
	})
	for _, want := range []string{"Giá thuê tối đa: 5.500.000 đồng", "Cọc tối đa: 1 tháng tiền thuê", "Tầng tối đa: 2", "Không giới hạn", "Nuôi thú cưng: Không yêu cầu", "Chỗ giữ xe: Bắt buộc", "Thang máy: Bắt buộc"} {
		if !strings.Contains(output, want) {
			t.Fatalf("profile view missing %q: %q", want, output)
		}
	}
	for _, raw := range []string{"true", "false", "VND"} {
		if strings.Contains(output, raw) {
			t.Fatalf("profile view contains raw value %q: %q", raw, output)
		}
	}
}

func TestProfileConfirmUsesVietnameseButtons(t *testing.T) {
	var rent, cash, floor, unknown string
	form := profileEditorForm(&preferences.SearchProfile{}, &rent, &cash, &floor, &unknown)
	form.Init()
	form.Update(tea.WindowSizeMsg{Width: 50, Height: 12})
	form.NextGroup()
	form.NextGroup()
	view := form.View()
	if !strings.Contains(view, "Có") || !strings.Contains(view, "Không") {
		t.Fatalf("confirmation buttons are missing: %q", view)
	}
	for _, english := range []string{"Yes", "No", "enter next"} {
		if strings.Contains(view, english) {
			t.Fatalf("confirmation contains English %q: %q", english, view)
		}
	}
}

func TestLegacyCashLimitPreservedUntilExplicitReplacement(t *testing.T) {
	legacy := int64(11_000_000)
	p := preferences.SearchProfile{MaxMoveInCashVND: &legacy, RequireParking: true}
	view := captureProfileOutput(t, func() { printProfile(p) })
	if !strings.Contains(view, "Giới hạn cũ") || !strings.Contains(view, "11.000.000 đồng") {
		t.Fatalf("legacy rule not explained: %q", view)
	}
	form := profileEditorForm(&p, new(string), new(string), new(string), new(string))
	form.Init()
	form.Update(tea.WindowSizeMsg{Width: 60, Height: 13})
	form.NextGroup()
	if !strings.Contains(form.View(), "thay thế") {
		t.Fatalf("legacy editor does not explain replacement: %q", form.View())
	}
	updated, err := profileFromInputs(p, "", "", "", "")
	if err != nil || updated.MaxMoveInCashVND == nil || updated.MaxDepositMonths != nil {
		t.Fatalf("unrelated edit removed legacy limit: %+v, %v", updated, err)
	}
	updated, err = profileFromInputs(p, "", "1", "", "")
	if err != nil || updated.MaxMoveInCashVND != nil || updated.MaxDepositMonths == nil || *updated.MaxDepositMonths != 1 {
		t.Fatalf("explicit replacement failed: %+v, %v", updated, err)
	}
}

func TestProfileMatchShowsDerivedMoveInCashAsReference(t *testing.T) {
	months := 1
	r := preferences.Evaluate(preferences.SearchProfile{MaxDepositMonths: &months}, preferences.RentalFacts{Price: "5tr", Deposit: "5tr"})
	output := captureProfileOutput(t, func() { RenderProfileMatch(r) })
	if !strings.Contains(output, "tham khảo") || !strings.Contains(output, "10.000.000 đồng") || !strings.Contains(output, "chưa tính khoản khác") {
		t.Fatalf("derived amount not qualified: %q", output)
	}
}

func TestProfileMatchUsesVietnameseLabels(t *testing.T) {
	rent := int64(5_000_000)
	r := preferences.Evaluate(preferences.SearchProfile{MaxRentVND: &rent, RequirePets: true}, preferences.RentalFacts{Price: "6tr"})
	output := captureProfileOutput(t, func() { RenderProfileMatch(r) })
	for _, want := range []string{"Không phù hợp", "Giá thuê (Không đạt)", "Nuôi thú cưng (Chưa rõ)", "Giá thuê", "chưa rõ"} {
		if !strings.Contains(output, want) {
			t.Fatalf("match view missing %q: %q", want, output)
		}
	}
	for _, raw := range []string{"Reject", "price", "pets_allowed", "(fail)", "(unknown)", "Rent ", "Pet permission"} {
		if strings.Contains(output, raw) {
			t.Fatalf("match view contains internal text %q: %q", raw, output)
		}
	}
}

func TestProfileUnknownLimitPromptExplainsMeaning(t *testing.T) {
	var rent, cash, floor, unknown string
	form := profileEditorForm(&preferences.SearchProfile{}, &rent, &cash, &floor, &unknown)
	form.Init()
	form.Update(tea.WindowSizeMsg{Width: 70, Height: 15})
	for range 6 {
		form.NextGroup()
	}
	view := form.View()
	if !strings.Contains(view, "Tin đăng được phép thiếu thông tin ở tối đa bao nhiêu tiêu chí?") ||
		!strings.Contains(view, "Thiếu hoặc không rõ đều được tính") {
		t.Fatalf("unknown limit prompt is unclear: %q", view)
	}
}

func TestProfileEditorEscapeReturnsWithoutAborting(t *testing.T) {
	rent := "5000000"
	var cash, floor, unknown string
	form := profileEditorForm(&preferences.SearchProfile{}, &rent, &cash, &floor, &unknown)
	m := &arrowModel{form: form}
	m.Init()
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.backPressed {
		t.Fatal("Esc did not return from the editor")
	}
	if form.State == huh.StateAborted {
		t.Fatal("Esc aborted the form instead of navigating back")
	}
}

func TestConfirmArrowsToggleWithoutNavigating(t *testing.T) {
	choice := false
	form := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("Cần chỗ giữ xe?").Value(&choice)))
	m := &arrowModel{form: form}
	m.Init()
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.backPressed || !choice {
		t.Fatalf("Left should toggle Yes/No, got choice=%t back=%t", choice, m.backPressed)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.backPressed || choice || form.State != huh.StateNormal {
		t.Fatalf("Right should toggle Yes/No without submitting, got choice=%t back=%t state=%v", choice, m.backPressed, form.State)
	}
}

func TestTextInputArrowsMoveCursorWithoutNavigating(t *testing.T) {
	value := "ab"
	form := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Giá thuê").Value(&value)))
	m := &arrowModel{form: form}
	m.Init()
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if m.backPressed || value != "aXb" || form.State != huh.StateNormal {
		t.Fatalf("Left insertion = %q, back=%t, state=%v", value, m.backPressed, form.State)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	if value != "aXbY" || form.State != huh.StateNormal {
		t.Fatalf("Right insertion = %q, state=%v", value, form.State)
	}
}

func TestMenuArrowsAndEnterSelect(t *testing.T) {
	var choice string
	selectField := huh.NewSelect[string]().Options(
		huh.NewOption("First", "first"),
		huh.NewOption("Second", "second"),
	).Value(&choice)
	form := huh.NewForm(huh.NewGroup(selectField))
	m := &arrowModel{form: form}
	m.Init()
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if hovered, _ := selectField.Hovered(); hovered != "second" || m.backPressed {
		t.Fatalf("Down hovered %q, back=%t", hovered, m.backPressed)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if hovered, _ := selectField.Hovered(); hovered != "first" {
		t.Fatalf("Up hovered %q", hovered)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if choice != "second" || m.backPressed {
		t.Fatalf("Enter chose %q, back=%t", choice, m.backPressed)
	}
}
