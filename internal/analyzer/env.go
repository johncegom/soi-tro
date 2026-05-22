package analyzer

import (
	"bufio"
	"os"
	"strings"
)

// ParseEnv parses raw string content representing a .env file format.
// It ignores empty lines and comment lines (starting with '#'),
// splits keys and values by the first '=', and strips wrapping quotes.
func ParseEnv(content string) map[string]string {
	envMap := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove wrapping quotes if present and string is long enough
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			if len(val) >= 2 {
				val = val[1 : len(val)-1]
			}
		}
		envMap[key] = val
	}
	return envMap
}

// LoadEnv reads a .env file from the specified path, parses its keys and values,
// and sets them as environment variables.
func LoadEnv(filename string) error {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	envMap := ParseEnv(string(bytes))
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	return nil
}
