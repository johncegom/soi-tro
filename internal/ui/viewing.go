package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"soi-tro/internal/database"
	"soi-tro/internal/gemini"
	"soi-tro/internal/viewing"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
)

const nextActionLayout = "2006-01-02"

// parseNextAction reads a YYYY-MM-DD calendar date. The date is kept as UTC
// midnight, the same zone the store round-trips, so the day never shifts with
// the machine's time zone. Blank input clears the date.
func parseNextAction(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	due, err := time.Parse(nextActionLayout, raw)
	if err != nil {
		return nil, errors.New("ngày không hợp lệ, dùng định dạng YYYY-MM-DD (ví dụ 2026-10-10)")
	}

	return &due, nil
}

// formatNextAction renders a next-action date as YYYY-MM-DD, or "" when unset.
func formatNextAction(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.UTC().Format(nextActionLayout)
}

func statusLabel(s viewing.DecisionStatus) string {
	switch s {
	case viewing.StatusUnreviewed:
		return "Chưa xem xét"
	case viewing.StatusViewingPlanned:
		return "Đã hẹn xem phòng"
	case viewing.StatusConsidering:
		return "Đang cân nhắc"
	case viewing.StatusRejected:
		return "Đã loại"
	case viewing.StatusReadyToNegotiate:
		return "Sẵn sàng thương lượng"
	default:
		return string(s)
	}
}

// trailSummary is the one-line status shown in the history list and comparison.
func trailSummary(t viewing.DecisionTrail) string {
	label := statusLabel(t.Status)
	if due := formatNextAction(t.NextAction); due != "" {
		return label + " | Việc tiếp theo: " + due
	}

	return label
}

// applyAnswer records an answer, or clears it when the text is blank.
func applyAnswer(t *viewing.DecisionTrail, id viewing.ItemID, text string) error {
	if strings.TrimSpace(text) == "" {
		t.ClearAnswer(id)

		return nil
	}

	return t.SetAnswer(id, text)
}

// claimsFromResult lists what the listing states, keyed by item ID. RawFields
// is the primary source; the standard fields fill in only where it lacks the
// key. Free-text notes are not a claim.
func claimsFromResult(r *gemini.RentalExtractionResult) map[viewing.ItemID]string {
	values := make(map[string]string, len(r.RawFields)+8)
	maps.Copy(values, r.RawFields)

	for k, v := range map[string]string{
		fieldPrice: r.Price, fieldDeposit: r.Deposit, fieldFloor: r.Floor, fieldElectricity: r.Electricity,
		fieldWater: r.Water, fieldParkingFee: r.ParkingFee, fieldPetsAllowed: r.PetsAllowed, keyPhoneNumber: r.PhoneNumber,
	} {
		if _, ok := values[k]; !ok {
			values[k] = v
		}
	}

	delete(values, "additional_notes")

	return viewing.ClaimsFromFields(values)
}

func itemLabel(id viewing.ItemID) string {
	if _, label, ok := strings.Cut(string(id), ":"); ok {
		return label
	}

	return string(id)
}

// renderReview prints listing claims, viewing answers and unresolved questions
// as three separate sections so a difference stays visible.
func renderReview(w io.Writer, t viewing.DecisionTrail, r viewing.Review) {
	var b strings.Builder

	fmt.Fprintf(&b, "\n📌 Phòng #%d - %s\n", t.RentalID, trailSummary(t))

	fmt.Fprintln(&b, "\n🏷️  Tin đăng (người đăng nói):")

	if len(r.Claims) == 0 {
		fmt.Fprintln(&b, "   (không có)")
	}

	for _, c := range r.Claims {
		fmt.Fprintf(&b, "   - %s: %s\n", itemLabel(c.ID), c.Claim)
	}

	fmt.Fprintln(&b, "\n📝 Câu trả lời khi xem phòng:")

	if len(r.Answers) == 0 {
		fmt.Fprintln(&b, "   (chưa có)")
	}

	for _, a := range r.Answers {
		fmt.Fprintf(&b, "   - %s: %s\n", itemLabel(a.ID), a.Answer)

		if a.Differs {
			fmt.Fprintf(&b, "     ⚠️  Khác tin đăng (%s)\n", a.Claim)
		}
	}

	fmt.Fprintln(&b, "\n❓ Chưa giải quyết:")

	if len(r.Unresolved) == 0 {
		fmt.Fprintln(&b, "   (đã trả lời hết)")
	}

	for _, it := range r.Unresolved {
		fmt.Fprintf(&b, "   - %s\n", it.Question)
	}

	fmt.Fprintln(&b)

	_, _ = io.WriteString(w, b.String())
}

// ViewingTrailUI lets the user pick a saved rental and work on its viewing
// checklist and decision trail.
func ViewingTrailUI() error {
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
		options = append(options, huh.NewOption(fmt.Sprintf("#%d - Giá: %s - ĐT: %s (%s)", rec.ID, rec.Result.Price, rec.Result.PhoneNumber, rec.CreatedAt), rec.ID))
	}

	var selectedID int64

	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[int64]().Title("Chọn phòng trọ để lập hồ sơ xem phòng").Options(options...).Value(&selectedID),
	))

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed {
		return err
	}

	for _, rec := range records {
		if rec.ID == selectedID {
			return trailMenu(rec)
		}
	}

	return nil
}

func trailMenu(rec database.RentalRecord) error {
	ctx := context.Background()
	store := database.NewDecisionTrailStore(database.DB)

	trail, err := store.Get(ctx, rec.ID)
	if err != nil {
		return err
	}

	checklist := viewing.GenerateChecklist(viewing.ChecklistInput{MissingFields: rec.Result.MissingFields})
	claims := claimsFromResult(rec.Result)

	for {
		var choice string

		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("HỒ SƠ XEM PHÒNG #%d - %s", rec.ID, trailSummary(trail))).
				Options(
					huh.NewOption("1. Ghi câu trả lời cho checklist", "answers"),
					huh.NewOption("2. Đổi trạng thái / ngày việc tiếp theo", "status"),
					huh.NewOption("3. Xem đối chiếu tin đăng - câu trả lời - câu hỏi còn mở", "review"),
					huh.NewOption("4. Soạn tin nhắn hỏi thêm", "followup"),
					huh.NewOption("5. Quay lại", backChoice),
				).
				Value(&choice),
		))

		backPressed, err := RunFormWithArrows(form)
		if err != nil {
			return err
		}

		if backPressed || choice == backChoice {
			return nil
		}

		switch choice {
		case "answers":
			err = editAnswers(ctx, store, &trail, viewing.EditableItems(checklist, claims))
		case "status":
			err = editStatus(ctx, store, &trail)
		case "review":
			renderReview(os.Stdout, trail, viewing.BuildReview(checklist, claims, trail))
		case "followup":
			err = followUp(viewing.BuildReview(checklist, claims, trail).Unresolved)
		}

		if err != nil {
			// Errors here never carry answer text; print and stay in the menu.
			fmt.Printf("❌ Lỗi: %v\n", err)
		}
	}
}

func editAnswers(ctx context.Context, store *database.DecisionTrailStore, trail *viewing.DecisionTrail, items []viewing.ChecklistItem) error {
	for {
		var options []huh.Option[viewing.ItemID]

		for _, it := range items {
			mark := "  "
			if _, ok := trail.Answers[it.ID]; ok {
				mark = "✔ "
			}

			options = append(options, huh.NewOption(mark+it.Question, it.ID))
		}

		var id viewing.ItemID

		pick := huh.NewForm(huh.NewGroup(
			huh.NewSelect[viewing.ItemID]().Title("Chọn câu hỏi để ghi câu trả lời (Esc để quay lại)").Options(options...).Value(&id),
		))

		backPressed, err := RunFormWithArrows(pick)
		if err != nil || backPressed {
			return err
		}

		text := trail.Answers[id]

		input := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Câu trả lời (để trống để xóa)").Value(&text),
		))

		backPressed, err = RunFormWithArrows(input)
		if err != nil {
			return err
		}

		if backPressed {
			continue
		}

		if err := applyAnswer(trail, id, text); err != nil {
			return err
		}

		if err := store.Save(ctx, *trail); err != nil {
			return err
		}
	}
}

func editStatus(ctx context.Context, store *database.DecisionTrailStore, trail *viewing.DecisionTrail) error {
	status := trail.Status
	due := formatNextAction(trail.NextAction)

	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[viewing.DecisionStatus]().Title("Trạng thái").Options(
			huh.NewOption(statusLabel(viewing.StatusUnreviewed), viewing.StatusUnreviewed),
			huh.NewOption(statusLabel(viewing.StatusViewingPlanned), viewing.StatusViewingPlanned),
			huh.NewOption(statusLabel(viewing.StatusConsidering), viewing.StatusConsidering),
			huh.NewOption(statusLabel(viewing.StatusRejected), viewing.StatusRejected),
			huh.NewOption(statusLabel(viewing.StatusReadyToNegotiate), viewing.StatusReadyToNegotiate),
		).Value(&status),
		huh.NewInput().Title("Ngày việc tiếp theo (YYYY-MM-DD, để trống để xóa)").Value(&due).Validate(func(s string) error {
			_, err := parseNextAction(s)

			return err
		}),
	))

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed {
		return err
	}

	next, err := parseNextAction(due)
	if err != nil {
		return err
	}

	if err := trail.SetStatus(status, next); err != nil {
		return err
	}

	return store.Save(ctx, *trail)
}

// followUp shows the questions still open and copies them only when the user
// says yes. The message holds questions only, never answers or notes.
func followUp(unresolved []viewing.ChecklistItem) error {
	msg := viewing.FollowUpMessage(unresolved)
	if msg == "" {
		fmt.Println("\n✅ Không còn câu hỏi nào để hỏi thêm.")

		return nil
	}

	fmt.Printf("\n%s\n\n", msg)

	var doCopy bool

	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Sao chép tin nhắn này vào clipboard?").Value(&doCopy),
	))

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed || !doCopy {
		return err
	}

	if err := clipboard.WriteAll(msg); err != nil {
		return fmt.Errorf("không thể sao chép vào clipboard: %w", err)
	}

	fmt.Println("📋 Đã sao chép tin nhắn vào clipboard.")

	return nil
}

// trailSummaryOrUnavailable lets history and comparison keep rendering when one
// trail cannot be read.
func trailSummaryOrUnavailable(t viewing.DecisionTrail, err error) string {
	if err != nil {
		return "không đọc được"
	}

	return trailSummary(t)
}

// trailSummaryFor reads (never writes) a rental's trail for display.
func trailSummaryFor(rentalID int64) string {
	trail, err := database.NewDecisionTrailStore(database.DB).Get(context.Background(), rentalID)

	return trailSummaryOrUnavailable(trail, err)
}
