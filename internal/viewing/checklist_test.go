package viewing

import "testing"

func ids(items []ChecklistItem) []ItemID {
	out := make([]ItemID, 0, len(items))
	for _, it := range items {
		out = append(out, it.ID)
	}

	return out
}

func indexOf(items []ChecklistItem, id ItemID) int {
	for i, it := range items {
		if it.ID == id {
			return i
		}
	}

	return -1
}

func TestGenerateChecklist_MissingFieldsComeFirstInInputOrder(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit", "pets_allowed"}})

	if len(items) < 2 || items[0].ID != "field:deposit" || items[1].ID != "field:pets_allowed" {
		t.Fatalf("want field:deposit then field:pets_allowed first, got %v", ids(items))
	}

	if items[0].Source != SourceMissingField {
		t.Errorf("Source = %v, want SourceMissingField", items[0].Source)
	}
}

func TestGenerateChecklist_AlwaysIncludesDefaultItems(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{})

	if len(items) == 0 {
		t.Fatal("empty input must still yield the default checklist")
	}

	for _, it := range items {
		if it.Source != SourceDefault {
			t.Errorf("item %q Source = %v, want SourceDefault", it.ID, it.Source)
		}
	}
}

func TestGenerateChecklist_DefaultsFollowMissingFields(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"water"}})

	field := indexOf(items, "field:water")
	if field != 0 {
		t.Fatalf("field:water index = %d, want 0", field)
	}

	if len(items) < 2 || items[1].Source != SourceDefault {
		t.Errorf("item after missing fields should be a default, got %v", ids(items))
	}
}

func TestGenerateChecklist_DeduplicatesAndSkipsBlank(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit", "", "  ", "deposit"}})

	seen := map[ItemID]int{}
	for _, it := range items {
		seen[it.ID]++
	}

	for id, n := range seen {
		if n != 1 {
			t.Errorf("item %q appears %d times, want 1", id, n)
		}
	}

	if seen["field:"] != 0 {
		t.Error("blank field key must not produce an item")
	}
}

func TestGenerateChecklist_UnknownFieldStillGetsQuestion(t *testing.T) {
	items := GenerateChecklist(ChecklistInput{MissingFields: []string{"balcony"}})

	if len(items) == 0 || items[0].ID != "field:balcony" || items[0].Question == "" {
		t.Errorf("unknown field must get a non-empty question, got %+v", items)
	}
}

func TestGenerateChecklist_IDsAreStableAcrossCalls(t *testing.T) {
	in := ChecklistInput{MissingFields: []string{"deposit"}}
	a, b := ids(GenerateChecklist(in)), ids(GenerateChecklist(in))

	if len(a) != len(b) {
		t.Fatalf("length differs: %d vs %d", len(a), len(b))
	}

	for i := range a {
		if a[i] != b[i] {
			t.Errorf("index %d: %q vs %q", i, a[i], b[i])
		}
	}
}
