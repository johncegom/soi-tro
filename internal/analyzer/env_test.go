package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			got := ParseEnv(tt.content)
			is.Equal(tt.expected, got)
		})
	}
}

func TestLoadEnv(t *testing.T) {
	// Test error when file does not exist
	err := LoadEnv("non_existent_file.env")
	assert.Error(t, err)

	// Create a temp file with .env content
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	content := "TEST_KEY_LOAD_ENV=test_value\nANOTHER_TEST_KEY=\"another_value\"\n"
	err = os.WriteFile(envFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Clean up environment variables afterwards
	t.Cleanup(func() {
		os.Unsetenv("TEST_KEY_LOAD_ENV")
		os.Unsetenv("ANOTHER_TEST_KEY")
	})

	// Load the env file
	err = LoadEnv(envFile)
	assert.NoError(t, err)

	// Verify env vars are set
	assert.Equal(t, "test_value", os.Getenv("TEST_KEY_LOAD_ENV"))
	assert.Equal(t, "another_value", os.Getenv("ANOTHER_TEST_KEY"))
}

func FuzzParseEnv(f *testing.F) {
	// Add seed corpus for fuzzing
	f.Add("KEY=VALUE")
	f.Add("# Comment\nKEY=\"Value\"")
	f.Add("KEY_ONE=val1\nKEY_TWO='val2'")
	f.Add("")

	f.Fuzz(func(t *testing.T, content string) {
		is := assert.New(t)
		// Parser must never panic regardless of the random content fuzzed
		got := ParseEnv(content)

		for k := range got {
			// Invariant check: Keys should never contain '='
			is.NotContains(k, "=")
		}
	})
}

// ExampleParseEnv shows how ParseEnv parses raw string content representing a .env file format.
func ExampleParseEnv() {
	content := "API_KEY=secret_key\nPORT=8080\n# comment\n"
	envMap := ParseEnv(content)
	fmt.Println(envMap["API_KEY"])
	fmt.Println(envMap["PORT"])

	// Output:
	// secret_key
	// 8080
}
