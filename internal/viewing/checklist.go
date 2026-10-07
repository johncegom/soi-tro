// Package viewing holds the pure rules behind viewing packs and the decision
// trail: checklist generation, answers, status transitions and follow-ups.
package viewing

import "strings"

// ItemID is the stable identifier of one checklist item.
type ItemID string

// ItemSource says why a checklist item exists.
type ItemSource int

const (
	// SourceMissingField marks an item generated from a field the listing omits.
	SourceMissingField ItemSource = iota
	// SourceDefault marks an item from the documented default checklist.
	SourceDefault
	// SourceVerify marks an editor-only item that checks a claim the listing made.
	SourceVerify
)

// ChecklistItem is one question to settle about a rental.
type ChecklistItem struct {
	ID       ItemID
	Question string
	Source   ItemSource
}

// ChecklistInput is everything GenerateChecklist needs.
type ChecklistInput struct {
	MissingFields []string
}

// fieldQuestions phrases the question for each standard field key. Any other
// key falls back to a generic question.
var fieldQuestions = map[string]string{
	"price":        "Giá thuê chính xác mỗi tháng là bao nhiêu?",
	"deposit":      "Tiền cọc là bao nhiêu và khi nào được hoàn lại?",
	"floor":        "Phòng ở tầng mấy, có thang máy không?",
	"electricity":  "Giá điện tính thế nào (đồng/số hay giá dân)?",
	"water":        "Giá nước tính thế nào (theo khối hay theo người)?",
	"parking_fee":  "Phí giữ xe là bao nhiêu mỗi tháng?",
	"pets_allowed": "Có được nuôi thú cưng không?",
	"phone_number": "Số điện thoại liên hệ của chủ nhà là gì?",
}

// defaultItems is the documented default viewing checklist, in display order.
var defaultItems = []ChecklistItem{
	{ID: "default:water-pressure", Question: "Nước có mạnh và nóng lạnh ổn định không?", Source: SourceDefault},
	{ID: "default:damp-mold", Question: "Tường, trần có ẩm mốc hoặc thấm dột không?", Source: SourceDefault},
	{ID: "default:noise", Question: "Tiếng ồn giờ cao điểm và ban đêm có chịu được không?", Source: SourceDefault},
	{ID: "default:fire-exit", Question: "Lối thoát hiểm và thiết bị chữa cháy nằm ở đâu?", Source: SourceDefault},
	{ID: "default:signal-internet", Question: "Sóng điện thoại và mạng internet trong phòng có ổn không?", Source: SourceDefault},
	{ID: "default:lease-terms", Question: "Thời hạn hợp đồng và điều kiện chấm dứt sớm là gì?", Source: SourceDefault},
}

// GenerateChecklist builds the viewing checklist for a rental: one item per
// distinct missing field in input order, followed by the default checklist.
func GenerateChecklist(input ChecklistInput) []ChecklistItem {
	items := make([]ChecklistItem, 0, len(input.MissingFields)+len(defaultItems))
	seen := make(map[ItemID]bool)

	for _, raw := range input.MissingFields {
		key := strings.TrimSpace(raw)
		if key == "" {
			continue
		}

		id := ItemID("field:" + key)
		if seen[id] {
			continue
		}

		seen[id] = true

		question, ok := fieldQuestions[key]
		if !ok {
			question = "Xác nhận thông tin còn thiếu: " + key + "?"
		}

		items = append(items, ChecklistItem{ID: id, Question: question, Source: SourceMissingField})
	}

	for _, it := range defaultItems {
		if !seen[it.ID] {
			seen[it.ID] = true
			items = append(items, it)
		}
	}

	return items
}
