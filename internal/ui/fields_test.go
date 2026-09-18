package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

func TestOrderedKeys_StandardFirstThenSortedCustom(t *testing.T) {
	props := map[string]*genai.Schema{}
	for _, k := range []string{"water", "zebra", "price", "air_conditioner", "missing_fields", "mango", "balcony", "elevator", "deposit"} {
		props[k] = &genai.Schema{}
	}
	standard := []string{"price", "deposit", "floor", "water"}

	want := []string{"price", "deposit", "water", "air_conditioner", "balcony", "elevator", "mango", "zebra"}

	// Map iteration is randomized, so repeat to catch unstable ordering.
	for i := 0; i < 50; i++ {
		assert.Equal(t, want, orderedKeys(props, standard, "missing_fields"))
	}
}

func TestOrderedKeys_NoSkipKeepsEverything(t *testing.T) {
	props := map[string]*genai.Schema{"b": {}, "a": {}, "price": {}}

	assert.Equal(t, []string{"price", "a", "b"}, orderedKeys(props, []string{"price"}))
}

func TestFieldMissing(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		missing  []string
		key      string
		value    string
		want     bool
	}{
		{"not required and empty", false, nil, "price", "", false},
		{"not required but listed missing", false, []string{"price"}, "price", "", false},
		{"required and present", true, nil, "price", "3tr", false},
		{"required and empty", true, nil, "price", "", true},
		{"required and blank", true, nil, "price", "  ", true},
		{"required n/a", true, nil, "price", "N/A", true},
		{"required không đề cập", true, nil, "price", " Không đề cập ", true},
		{"required chưa đề cập", true, nil, "price", "chưa đề cập", true},
		{"required and listed missing despite value", true, []string{"PRICE"}, "price", "3tr", true},
		{"required, other key listed missing", true, []string{"deposit"}, "price", "3tr", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fieldMissing(tt.required, tt.missing, tt.key, tt.value))
		})
	}
}
