package viewing

import (
	"slices"
	"strings"
)

// ClaimLine is one thing the listing states.
type ClaimLine struct {
	ID    ItemID
	Claim string
}

// AnswerLine is one viewing answer, with the listing claim it can be checked
// against. Differs is true when a claim exists and the answer does not match it.
type AnswerLine struct {
	ID      ItemID
	Answer  string
	Claim   string
	Differs bool
}

// Review keeps listing claims, viewing answers and unresolved questions apart
// so a difference stays visible instead of being overwritten.
type Review struct {
	Claims     []ClaimLine
	Answers    []AnswerLine
	Unresolved []ChecklistItem
}

// ClaimsFromFields turns extracted listing values (field key to value) into
// claims keyed by item ID, skipping absent and "not mentioned" values.
func ClaimsFromFields(values map[string]string) map[ItemID]string {
	claims := make(map[ItemID]string)

	for key, value := range values {
		value = strings.TrimSpace(value)
		if isAbsent(value) {
			continue
		}

		claims[ItemID("field:"+key)] = value
	}

	return claims
}

// isAbsent reports whether a listing value says nothing. It mirrors the
// placeholders the Gemini layer treats as missing.
func isAbsent(value string) bool {
	switch strings.ToLower(value) {
	case "", "n/a", "không đề cập", "chưa đề cập":
		return true
	default:
		return false
	}
}

// sameText compares two values ignoring case and runs of whitespace.
func sameText(a, b string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(a), " "), strings.Join(strings.Fields(b), " "))
}

func sortedIDs[V any](m map[ItemID]V) []ItemID {
	out := make([]ItemID, 0, len(m))
	for id := range m {
		out = append(out, id)
	}

	slices.Sort(out)

	return out
}

// BuildReview groups a checklist, the listing's claims and a trail into a
// Review. Claims and answers are ordered by item ID; unresolved items keep
// checklist order.
func BuildReview(items []ChecklistItem, claims map[ItemID]string, trail DecisionTrail) Review {
	var r Review

	for _, id := range sortedIDs(claims) {
		r.Claims = append(r.Claims, ClaimLine{ID: id, Claim: claims[id]})
	}

	for _, id := range sortedIDs(trail.Answers) {
		answer, claim := trail.Answers[id], claims[id]
		r.Answers = append(r.Answers, AnswerLine{
			ID:      id,
			Answer:  answer,
			Claim:   claim,
			Differs: claim != "" && !sameText(claim, answer),
		})
	}

	for _, it := range items {
		if _, answered := trail.Answers[it.ID]; !answered {
			r.Unresolved = append(r.Unresolved, it)
		}
	}

	return r
}
