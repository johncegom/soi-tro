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
	t.Run("file does not exist", func(t *testing.T) {
		err := LoadEnv("non_existent_file.env")
		assert.Error(t, err)
	})

	t.Run("successful load with various formats", func(t *testing.T) {
		// Create a temp file with .env content
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := "TEST_KEY_LOAD_ENV=test_value\nANOTHER_TEST_KEY=\"another_value\"\n"
		err := os.WriteFile(envFile, []byte(content), 0644)
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
	})

	t.Run("load with comments and empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := "# This is a comment\n\nKEY1=value1\n  \nKEY2=value2\n"
		err := os.WriteFile(envFile, []byte(content), 0644)
		assert.NoError(t, err)

		t.Cleanup(func() {
			os.Unsetenv("KEY1")
			os.Unsetenv("KEY2")
		})

		err = LoadEnv(envFile)
		assert.NoError(t, err)
		assert.Equal(t, "value1", os.Getenv("KEY1"))
		assert.Equal(t, "value2", os.Getenv("KEY2"))
	})

	t.Run("load with spaces around equals", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := "  KEY_ONE  =   val1  \nKEY_TWO = val2"
		err := os.WriteFile(envFile, []byte(content), 0644)
		assert.NoError(t, err)

		t.Cleanup(func() {
			os.Unsetenv("KEY_ONE")
			os.Unsetenv("KEY_TWO")
		})

		err = LoadEnv(envFile)
		assert.NoError(t, err)
		assert.Equal(t, "val1", os.Getenv("KEY_ONE"))
		assert.Equal(t, "val2", os.Getenv("KEY_TWO"))
	})

	t.Run("load with quoted values", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := "DOUBLE_QUOTED=\"value with spaces\"\nSINGLE_QUOTED='another value'"
		err := os.WriteFile(envFile, []byte(content), 0644)
		assert.NoError(t, err)

		t.Cleanup(func() {
			os.Unsetenv("DOUBLE_QUOTED")
			os.Unsetenv("SINGLE_QUOTED")
		})

		err = LoadEnv(envFile)
		assert.NoError(t, err)
		assert.Equal(t, "value with spaces", os.Getenv("DOUBLE_QUOTED"))
		assert.Equal(t, "another value", os.Getenv("SINGLE_QUOTED"))
	})

	t.Run("load with invalid lines ignored", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := "VALID_KEY=valid_value\nINVALID_LINE_NO_EQUALS\nANOTHER_VALID=another"
		err := os.WriteFile(envFile, []byte(content), 0644)
		assert.NoError(t, err)

		t.Cleanup(func() {
			os.Unsetenv("VALID_KEY")
			os.Unsetenv("ANOTHER_VALID")
		})

		err = LoadEnv(envFile)
		assert.NoError(t, err)
		assert.Equal(t, "valid_value", os.Getenv("VALID_KEY"))
		assert.Equal(t, "another", os.Getenv("ANOTHER_VALID"))
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := ""
		err := os.WriteFile(envFile, []byte(content), 0644)
		assert.NoError(t, err)

		err = LoadEnv(envFile)
		assert.NoError(t, err)
	})
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
	content := "API_KEY=mock_key\nPORT=8080\n# comment\n"
	envMap := ParseEnv(content)
	fmt.Println(envMap["API_KEY"])
	fmt.Println(envMap["PORT"])

	// Output:
	// mock_key
	// 8080
}
