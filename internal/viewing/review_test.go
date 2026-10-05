package viewing

import (
	"strings"
	"testing"
)

func TestClaimsFromFields_SkipsAbsentAndPlaceholders(t *testing.T) {
	got := ClaimsFromFields(map[string]string{
		"price":   "4,500,000 VND",
		"deposit": "Không đề cập",
		"floor":   "  ",
		"water":   "N/A",
		"parking": "Chưa đề cập",
		"pets":    " Cho nuôi mèo ",
	})

	if len(got) != 2 || got["field:price"] != "4,500,000 VND" || got["field:pets"] != "Cho nuôi mèo" {
		t.Errorf("claims = %v, want only field:price and field:pets (trimmed)", got)
	}
}

func TestBuildReview_GroupsAreSeparate(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit"}})
	claims := map[ItemID]string{"field:price": "4,500,000 VND"}
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:deposit", "1 tháng")

	r := BuildReview(items, claims, tr)

	if len(r.Claims) != 1 || r.Claims[0].ID != "field:price" || r.Claims[0].Claim != "4,500,000 VND" {
		t.Errorf("Claims = %+v", r.Claims)
	}

	if len(r.Answers) != 1 || r.Answers[0].ID != "field:deposit" || r.Answers[0].Answer != "1 tháng" {
		t.Errorf("Answers = %+v", r.Answers)
	}

	if indexOf(r.Unresolved, "field:deposit") != -1 {
		t.Error("answered item must not be unresolved")
	}

	if indexOf(r.Unresolved, "default:noise") == -1 {
		t.Error("unanswered default item must be unresolved")
	}
}

func TestBuildReview_UnresolvedKeepsChecklistOrder(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit", "water"}})
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:deposit", "1 tháng")

	r := BuildReview(items, nil, tr)

	if len(r.Unresolved) == 0 || r.Unresolved[0].ID != "field:water" {
		t.Errorf("first unresolved = %v, want field:water", ids(r.Unresolved))
	}
}

func TestBuildReview_DiffersWhenAnswerContradictsClaim(t *testing.T) {
	claims := map[ItemID]string{"field:price": "4,500,000 VND"}
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:price", "5,000,000 VND")

	r := BuildReview(nil, claims, tr)

	if len(r.Answers) != 1 || !r.Answers[0].Differs {
		t.Fatalf("Answers = %+v, want Differs=true", r.Answers)
	}

	if r.Answers[0].Claim != "4,500,000 VND" || r.Answers[0].Answer != "5,000,000 VND" {
		t.Errorf("both values must stay visible: %+v", r.Answers[0])
	}
}

func TestBuildReview_NoDifferenceWhenOnlyCaseOrSpacingDiffers(t *testing.T) {
	claims := map[ItemID]string{"field:pets": "Cho nuôi mèo"}
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:pets", "  cho   NUÔI mèo ")

	r := BuildReview(nil, claims, tr)

	if len(r.Answers) != 1 || r.Answers[0].Differs {
		t.Errorf("Answers = %+v, want Differs=false", r.Answers)
	}
}

func TestBuildReview_NoDifferenceWithoutClaim(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("default:noise", "ồn")

	r := BuildReview(nil, nil, tr)

	if len(r.Answers) != 1 || r.Answers[0].Differs || r.Answers[0].Claim != "" {
		t.Errorf("Answers = %+v, want Differs=false and empty claim", r.Answers)
	}
}

func TestBuildReview_AnswersAreDeterministicallyOrdered(t *testing.T) {
	tr := NewTrail(1)
	_ = tr.SetAnswer("field:b", "2")
	_ = tr.SetAnswer("field:a", "1")
	_ = tr.SetAnswer("field:c", "3")

	for range 20 {
		r := BuildReview(nil, nil, tr)

		var got []string
		for _, a := range r.Answers {
			got = append(got, string(a.ID))
		}

		if strings.Join(got, ",") != "field:a,field:b,field:c" {
			t.Fatalf("order = %v, want sorted by ID", got)
		}
	}
}

func TestBuildReview_ClaimsAreDeterministicallyOrdered(t *testing.T) {
	claims := map[ItemID]string{"field:b": "2", "field:a": "1"}

	for range 20 {
		r := BuildReview(nil, claims, NewTrail(1))

		if len(r.Claims) != 2 || r.Claims[0].ID != "field:a" || r.Claims[1].ID != "field:b" {
			t.Fatalf("Claims = %+v, want sorted by ID", r.Claims)
		}
	}
}
