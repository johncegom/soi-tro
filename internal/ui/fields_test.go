package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

const (
	keyAlpha = "alpha"
	keyRent  = "rent"
)

func TestOrderedKeys_StandardFirstThenSortedCustom(t *testing.T) {
	t.Parallel()

	props := map[string]*genai.Schema{}
	for _, k := range []string{"beta", "zebra", "air_conditioner", "skipme", "mango", "balcony", "elevator", keyAlpha} {
		props[k] = &genai.Schema{}
	}

	standard := []string{keyAlpha, "beta"}

	want := []string{keyAlpha, "beta", "air_conditioner", "balcony", "elevator", "mango", "zebra"}

	// Map iteration is randomized, so repeat to catch unstable ordering.
	for range 50 {
		assert.Equal(t, want, orderedKeys(props, standard, "skipme"))
	}
}

func TestOrderedKeys_NoSkipKeepsEverything(t *testing.T) {
	t.Parallel()

	props := map[string]*genai.Schema{"b": {}, "a": {}, keyAlpha: {}}

	assert.Equal(t, []string{keyAlpha, "a", "b"}, orderedKeys(props, []string{keyAlpha}))
}

func TestFieldMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		required bool
		missing  []string
		key      string
		value    string
		want     bool
	}{
		{"not required and empty", false, nil, keyRent, "", false},
		{"not required but listed missing", false, []string{keyRent}, keyRent, "", false},
		{"required and present", true, nil, keyRent, "3tr", false},
		{"required and empty", true, nil, keyRent, "", true},
		{"required and blank", true, nil, keyRent, "  ", true},
		{"required n/a", true, nil, keyRent, "N/A", true},
		{"required không đề cập", true, nil, keyRent, " Không đề cập ", true},
		{"required chưa đề cập", true, nil, keyRent, "chưa đề cập", true},
		{"required and listed missing despite value", true, []string{"RENT"}, keyRent, "3tr", true},
		{"required, other key listed missing", true, []string{"fee"}, keyRent, "3tr", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, fieldMissing(tt.required, tt.missing, tt.key, tt.value))
		})
	}
}
