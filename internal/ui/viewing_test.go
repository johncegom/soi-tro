package ui

import (
	"bytes"
	"errors"
	"soi-tro/internal/gemini"
	"soi-tro/internal/viewing"
	"strings"
	"testing"
	"time"
)

func TestParseNextActionBlankClears(t *testing.T) {
	got, err := parseNextAction("  ")
	if err != nil || got != nil {
		t.Fatalf("blank = (%v, %v), want (nil, nil)", got, err)
	}
}

func TestParseNextActionRejectsBadDate(t *testing.T) {
	for _, raw := range []string{"10/10/2026", "2026-13-01", "tomorrow"} {
		if got, err := parseNextAction(raw); err == nil {
			t.Fatalf("parseNextAction(%q) = %v, want error", raw, got)
		}
	}
}

func TestNextActionDayDoesNotShiftWithLocalZone(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("UTC+7", 7*60*60)
	defer func() { time.Local = old }()

	due, err := parseNextAction("2026-10-10")
	if err != nil {
		t.Fatal(err)
	}

	// The store keeps UTC, so round-trip through UTC the way Get returns it.
	roundTripped := due.UTC()
	if got := formatNextAction(&roundTripped); got != "2026-10-10" {
		t.Fatalf("formatNextAction = %q, want 2026-10-10", got)
	}
}

func TestFormatNextActionNilIsEmpty(t *testing.T) {
	if got := formatNextAction(nil); got != "" {
		t.Fatalf("formatNextAction(nil) = %q, want empty", got)
	}
}

func TestStatusLabelCoversEveryStatus(t *testing.T) {
	statuses := []viewing.DecisionStatus{
		viewing.StatusUnreviewed, viewing.StatusViewingPlanned, viewing.StatusConsidering,
		viewing.StatusRejected, viewing.StatusReadyToNegotiate,
	}
	seen := map[string]bool{}

	for _, s := range statuses {
		label := statusLabel(s)
		if label == "" || label == string(s) || seen[label] {
			t.Fatalf("statusLabel(%q) = %q: want a distinct Vietnamese label", s, label)
		}

		seen[label] = true
	}
}

func TestApplyAnswerBlankClears(t *testing.T) {
	trail := viewing.NewTrail(1)
	if err := applyAnswer(&trail, "default:noise", "yên"); err != nil {
		t.Fatal(err)
	}

	if err := applyAnswer(&trail, "default:noise", "   "); err != nil {
		t.Fatalf("blank answer returned %v, want it to clear", err)
	}

	if _, ok := trail.Answers["default:noise"]; ok {
		t.Fatal("blank answer did not clear the stored answer")
	}
}

func TestApplyAnswerTrimsAndStores(t *testing.T) {
	trail := viewing.NewTrail(1)
	if err := applyAnswer(&trail, "default:noise", "  yên  "); err != nil {
		t.Fatal(err)
	}

	if got := trail.Answers["default:noise"]; got != "yên" {
		t.Fatalf("answer = %q, want yên", got)
	}
}

func TestClaimsFromResultPrefersRawFieldsAndDropsNotes(t *testing.T) {
	r := &gemini.RentalExtractionResult{
		Price:           "4tr5",
		Deposit:         "1 tháng",
		AdditionalNotes: "long free text",
		RawFields:       map[string]string{"price": "4.500.000", "wifi": "free", "additional_notes": "long free text"},
	}

	claims := claimsFromResult(r)

	if claims["field:price"] != "4.500.000" {
		t.Fatalf("price = %q, want the RawFields value", claims["field:price"])
	}

	if claims["field:deposit"] != "1 tháng" {
		t.Fatalf("deposit = %q, want the struct value filled in", claims["field:deposit"])
	}

	if claims["field:wifi"] != "free" {
		t.Fatalf("custom field missing: %v", claims)
	}

	if _, ok := claims["field:additional_notes"]; ok {
		t.Fatal("additional_notes must not be a claim")
	}
}

func TestClaimsFromResultSkipsAbsentValues(t *testing.T) {
	r := &gemini.RentalExtractionResult{Price: "Không đề cập"}

	if claims := claimsFromResult(r); len(claims) != 0 {
		t.Fatalf("claims = %v, want none", claims)
	}
}

func TestTrailSummary(t *testing.T) {
	trail := viewing.NewTrail(1)
	if got := trailSummary(trail); got != statusLabel(viewing.StatusUnreviewed) {
		t.Fatalf("summary = %q, want just the status label", got)
	}

	due := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	if err := trail.SetStatus(viewing.StatusViewingPlanned, &due); err != nil {
		t.Fatal(err)
	}

	got := trailSummary(trail)
	if !strings.Contains(got, statusLabel(viewing.StatusViewingPlanned)) || !strings.Contains(got, "2026-10-10") {
		t.Fatalf("summary = %q, want status label and date", got)
	}
}

func TestRenderReviewKeepsThreeSectionsApart(t *testing.T) {
	items := viewing.GenerateChecklist(viewing.ChecklistInput{})
	claims := map[viewing.ItemID]string{"field:price": "4tr5"}
	trail := viewing.NewTrail(7)
	_ = trail.SetAnswer("field:price", "5tr")
	_ = trail.SetAnswer("default:noise", "yên tĩnh")

	var buf bytes.Buffer
	renderReview(&buf, trail, viewing.BuildReview(items, claims, trail))
	out := buf.String()

	for _, want := range []string{"Tin đăng", "4tr5", "Câu trả lời", "5tr", "yên tĩnh", "Chưa giải quyết", "Khác tin đăng"} {
		if !strings.Contains(out, want) {
			t.Fatalf("review missing %q:\n%s", want, out)
		}
	}

	if strings.Contains(out, items[0].Question) && !strings.Contains(out, "Chưa giải quyết") {
		t.Fatal("unresolved questions lack their own section")
	}
}

func TestRenderReviewDoesNotFlagMatchingAnswer(t *testing.T) {
	claims := map[viewing.ItemID]string{"field:price": "4tr5"}
	trail := viewing.NewTrail(7)
	_ = trail.SetAnswer("field:price", "4TR5")

	var buf bytes.Buffer
	renderReview(&buf, trail, viewing.BuildReview(nil, claims, trail))

	if strings.Contains(buf.String(), "Khác tin đăng") {
		t.Fatalf("matching answer flagged as different:\n%s", buf.String())
	}
}

func TestTrailSummaryOrUnavailable(t *testing.T) {
	trail := viewing.NewTrail(1)

	if got := trailSummaryOrUnavailable(trail, nil); got != statusLabel(viewing.StatusUnreviewed) {
		t.Fatalf("ok summary = %q", got)
	}

	if got := trailSummaryOrUnavailable(trail, errors.New("db down")); got != "không đọc được" {
		t.Fatalf("error summary = %q, want không đọc được", got)
	}
}
