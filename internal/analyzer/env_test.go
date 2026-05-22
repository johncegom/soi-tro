package analyzer

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseEnv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:    "simple key value",
			content: "KEY=VALUE",
			expected: map[string]string{
				"KEY": "VALUE",
			},
		},
		{
			name:    "ignores comments and empty lines",
			content: "\n# This is a comment\n\nKEY=VALUE\n   # another comment\n",
			expected: map[string]string{
				"KEY": "VALUE",
			},
		},
		{
			name:    "double quotes stripped",
			content: `KEY="VALUE_WITH_QUOTES"`,
			expected: map[string]string{
				"KEY": "VALUE_WITH_QUOTES",
			},
		},
		{
			name:    "single quotes stripped",
			content: "KEY='VALUE_WITH_QUOTES'",
			expected: map[string]string{
				"KEY": "VALUE_WITH_QUOTES",
			},
		},
		{
			name:    "multiple values and spaces",
			content: "  KEY_ONE  =   val1  \nKEY_TWO=\"val2\"",
			expected: map[string]string{
				"KEY_ONE": "val1",
				"KEY_TWO": "val2",
			},
		},
		{
			name:     "invalid format ignored",
			content:  "INVALID_LINE_NO_EQUALS",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseEnv(tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseEnv() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func FuzzParseEnv(f *testing.F) {
	// Add seed corpus for fuzzing
	f.Add("KEY=VALUE")
	f.Add("# Comment\nKEY=\"Value\"")
	f.Add("KEY_ONE=val1\nKEY_TWO='val2'")
	f.Add("")

	f.Fuzz(func(t *testing.T, content string) {
		// Parser must never panic regardless of the random content fuzzed
		got := ParseEnv(content)

		for k := range got {
			// Invariant check: Keys should never contain '='
			if strings.Contains(k, "=") {
				t.Errorf("Parsed key %q contains '='", k)
			}
		}
	})
}
