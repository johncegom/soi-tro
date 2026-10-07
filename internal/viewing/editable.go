package viewing

import "strings"

// EditableItems returns the items a user can answer at a viewing: the checklist
// followed by one verify item per claimed field the checklist does not already
// cover, ordered by ID. Answering a verify item lets a viewing answer be checked
// against the listing's claim. Verify items are not checklist questions, so they
// never reach the follow-up message.
func EditableItems(checklist []ChecklistItem, claims map[ItemID]string) []ChecklistItem {
	items := make([]ChecklistItem, 0, len(checklist)+len(claims))
	items = append(items, checklist...)

	seen := make(map[ItemID]bool, len(checklist))
	for _, it := range checklist {
		seen[it.ID] = true
	}

	for _, id := range sortedIDs(claims) {
		if seen[id] {
			continue
		}

		key := strings.TrimPrefix(string(id), "field:")
		items = append(items, ChecklistItem{
			ID:       id,
			Question: "Xác nhận: " + key + " (tin đăng: " + claims[id] + ")",
			Source:   SourceVerify,
		})
	}

	return items
}
