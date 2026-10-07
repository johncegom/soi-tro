package viewing

import "testing"

func TestEditableItemsAddsVerifyItemsForClaimedFields(t *testing.T) {
	checklist := GenerateChecklist(ChecklistInput{MissingFields: []string{"deposit"}})
	claims := map[ItemID]string{"field:price": "4tr5", "field:deposit": "1 tháng"}

	items := EditableItems(checklist, claims)

	if len(items) != len(checklist)+1 {
		t.Fatalf("got %d items, want checklist plus one verify item", len(items))
	}

	for i, it := range checklist {
		if items[i] != it {
			t.Fatalf("item %d changed: got %+v, want %+v", i, items[i], it)
		}
	}

	verify := items[len(items)-1]
	if verify.ID != "field:price" {
		t.Fatalf("verify item ID = %q, want field:price", verify.ID)
	}

	if verify.Source != SourceVerify {
		t.Fatalf("verify source = %v, want SourceVerify", verify.Source)
	}

	if want := "Xác nhận: price (tin đăng: 4tr5)"; verify.Question != want {
		t.Fatalf("verify question = %q, want %q", verify.Question, want)
	}
}

func TestEditableItemsOrdersVerifyItemsByID(t *testing.T) {
	claims := map[ItemID]string{"field:water": "100k", "field:floor": "3"}

	items := EditableItems(nil, claims)

	if len(items) != 2 || items[0].ID != "field:floor" || items[1].ID != "field:water" {
		t.Fatalf("verify items not ordered by ID: %+v", items)
	}
}

func TestEditableItemsDoesNotChangeFollowUp(t *testing.T) {
	checklist := GenerateChecklist(ChecklistInput{})
	claims := map[ItemID]string{"field:price": "4tr5"}

	_ = EditableItems(checklist, claims)

	review := BuildReview(checklist, claims, NewTrail(1))
	for _, it := range review.Unresolved {
		if it.ID == "field:price" {
			t.Fatal("verify item leaked into unresolved questions")
		}
	}
}
