package viewing

import (
	"errors"
	"testing"
	"time"
)

func TestNewTrail_StartsUnreviewedAndEmpty(t *testing.T) {
	tr := NewTrail(42)

	if tr.RentalID != 42 || tr.Status != StatusUnreviewed || tr.NextAction != nil || len(tr.Answers) != 0 {
		t.Errorf("NewTrail(42) = %+v", tr)
	}
}

func TestDecisionStatus_Valid(t *testing.T) {
	for _, s := range []DecisionStatus{
		StatusUnreviewed, StatusViewingPlanned, StatusConsidering, StatusRejected, StatusReadyToNegotiate,
	} {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}

	for _, s := range []DecisionStatus{"", "done", "REJECTED"} {
		if s.Valid() {
			t.Errorf("%q should be invalid", s)
		}
	}
}

func TestSetStatus_AnyValidToAnyValid(t *testing.T) {
	tr := NewTrail(1)

	for _, s := range []DecisionStatus{StatusReadyToNegotiate, StatusRejected, StatusConsidering, StatusUnreviewed} {
		if err := tr.SetStatus(s, nil); err != nil || tr.Status != s {
			t.Fatalf("SetStatus(%q): err=%v status=%q", s, err, tr.Status)
		}
	}
}

func TestSetStatus_InvalidLeavesTrailUnchanged(t *testing.T) {
	tr := NewTrail(1)
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	if err := tr.SetStatus(StatusConsidering, &due); err != nil {
		t.Fatal(err)
	}

	err := tr.SetStatus("bogus", nil)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("err = %v, want ErrInvalidStatus", err)
	}

	if tr.Status != StatusConsidering || tr.NextAction == nil || !tr.NextAction.Equal(due) {
		t.Errorf("trail changed on invalid status: %+v", tr)
	}
}

func TestSetStatus_NextActionSetAndCleared(t *testing.T) {
	tr := NewTrail(1)
	due := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

	if err := tr.SetStatus(StatusViewingPlanned, &due); err != nil {
		t.Fatal(err)
	}

	if tr.NextAction == nil || !tr.NextAction.Equal(due) {
		t.Fatalf("NextAction = %v, want %v", tr.NextAction, due)
	}

	if err := tr.SetStatus(StatusViewingPlanned, nil); err != nil {
		t.Fatal(err)
	}

	if tr.NextAction != nil {
		t.Errorf("NextAction = %v, want nil after clearing", tr.NextAction)
	}
}

func TestSetStatus_CopiesNextAction(t *testing.T) {
	tr := NewTrail(1)
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	if err := tr.SetStatus(StatusConsidering, &due); err != nil {
		t.Fatal(err)
	}

	due = due.Add(24 * time.Hour)

	if tr.NextAction == nil || tr.NextAction.Equal(due) {
		t.Error("trail must not alias the caller's time value")
	}
}

func TestSetAnswer_RecordEditAndTrim(t *testing.T) {
	tr := NewTrail(1)

	if err := tr.SetAnswer("field:deposit", "  1 tháng  "); err != nil {
		t.Fatal(err)
	}

	if got := tr.Answers["field:deposit"]; got != "1 tháng" {
		t.Errorf("answer = %q, want trimmed %q", got, "1 tháng")
	}

	if err := tr.SetAnswer("field:deposit", "2 tháng"); err != nil {
		t.Fatal(err)
	}

	if got := tr.Answers["field:deposit"]; got != "2 tháng" {
		t.Errorf("edited answer = %q, want %q", got, "2 tháng")
	}
}

func TestSetAnswer_BlankRejectedAndKeepsExisting(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("default:noise", "ồn")

	for _, blank := range []string{"", "   ", "\n\t"} {
		if err := tr.SetAnswer("default:noise", blank); !errors.Is(err, ErrEmptyAnswer) {
			t.Errorf("SetAnswer(%q) err = %v, want ErrEmptyAnswer", blank, err)
		}
	}

	if tr.Answers["default:noise"] != "ồn" {
		t.Errorf("existing answer lost: %q", tr.Answers["default:noise"])
	}
}

func TestClearAnswer(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("default:noise", "ồn")

	tr.ClearAnswer("default:noise")
	tr.ClearAnswer("never-set")

	if _, ok := tr.Answers["default:noise"]; ok {
		t.Error("answer should be gone after ClearAnswer")
	}
}

func TestAnswersSurviveChecklistRegeneration(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:deposit", "1 tháng")

	_ = GenerateChecklist(ChecklistInput{MissingFields: []string{"water"}})

	if tr.Answers["field:deposit"] != "1 tháng" {
		t.Error("answer for an item no longer in the checklist must be retained")
	}
}
