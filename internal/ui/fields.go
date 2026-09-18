package ui

import (
	"slices"
	"strings"

	"google.golang.org/genai"
)

// standardFieldKeys is the display order of the built-in fields; custom fields follow.
var standardFieldKeys = []string{"price", "deposit", "floor", "parking_fee", "pets_allowed", "electricity", "water"}

// nonDisplayKeys are schema properties shown outside the field table.
var nonDisplayKeys = []string{"missing_fields", "sample_messages", "additional_notes", "phone_number"}

// orderedKeys returns the property keys with the standard keys first (in the given order),
// then the remaining keys sorted, omitting skip. Map iteration is random, so sorting keeps
// custom-field rows stable across runs.
func orderedKeys(props map[string]*genai.Schema, standard []string, skip ...string) []string {
	keys := make([]string, 0, len(props))

	for _, k := range standard {
		if _, ok := props[k]; ok && !slices.Contains(skip, k) {
			keys = append(keys, k)
		}
	}

	var rest []string

	for k := range props {
		if !slices.Contains(standard, k) && !slices.Contains(skip, k) {
			rest = append(rest, k)
		}
	}

	slices.Sort(rest)

	return append(keys, rest...)
}

// fieldMissing reports whether a required field lacks an answer: the model listed it as
// missing, or its value is blank or a "not mentioned" placeholder.
func fieldMissing(required bool, missing []string, key, value string) bool {
	if !required {
		return false
	}

	for _, m := range missing {
		if strings.EqualFold(m, key) {
			return true
		}
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "n/a", "không đề cập", "chưa đề cập":
		return true
	}

	return false
}
