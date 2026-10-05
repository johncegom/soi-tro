package viewing

import (
	"strings"
	"testing"
)

func TestFollowUpMessage_ContainsOnlyGivenQuestions(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit", "water"}})
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:deposit", "1 tháng")

	r := BuildReview(items, nil, tr)
	msg := FollowUpMessage(r.Unresolved)

	if strings.Contains(msg, fieldQuestions["deposit"]) {
		t.Error("answered question must not appear in the follow-up")
	}

	if !strings.Contains(msg, fieldQuestions["water"]) {
		t.Error("unresolved question must appear in the follow-up")
	}
}

func TestFollowUpMessage_NumbersQuestionsInOrder(t *testing.T) {
	msg := FollowUpMessage([]ChecklistItem{
		{ID: "field:a", Question: "Câu một?"},
		{ID: "field:b", Question: "Câu hai?"},
	})

	one, two := strings.Index(msg, "1. Câu một?"), strings.Index(msg, "2. Câu hai?")
	if one < 0 || two < 0 || one > two {
		t.Errorf("message = %q, want numbered questions in order", msg)
	}
}

func TestFollowUpMessage_EmptyWhenNothingUnresolved(t *testing.T) {
	if got := FollowUpMessage(nil); got != "" {
		t.Errorf("FollowUpMessage(nil) = %q, want empty", got)
	}
}

func TestFollowUpMessage_DoesNotLeakAnswers(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("default:noise", "ghi chú riêng tư SECRET")

	items := GenerateChecklist(ChecklistInput{})
	msg := FollowUpMessage(BuildReview(items, nil, tr).Unresolved)

	if strings.Contains(msg, "SECRET") {
		t.Error("follow-up must never include the user's answers or notes")
	}
}

func TestFollowUpMessage_UsesAgreedPronouns(t *testing.T) {
	msg := FollowUpMessage([]ChecklistItem{{ID: "field:a", Question: "Câu một?"}})

	for _, banned := range []string{" em ", " anh ", " chị ", " tôi "} {
		if strings.Contains(strings.ToLower(msg), banned) {
			t.Errorf("message uses banned pronoun %q: %q", banned, msg)
		}
	}

	if !strings.Contains(msg, "mình") {
		t.Errorf("message should address the sender as \"mình\": %q", msg)
	}
}
