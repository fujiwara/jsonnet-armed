package functions

import (
	"strings"
	"testing"
)

func TestGenerateArmedLib(t *testing.T) {
	result := GenerateArmedLib()

	// Verify it contains expected function definitions
	expectedFunctions := []string{
		"env: std.native('env')",
		"sha256: std.native('sha256')",
		"file_content: std.native('file_content')",
		"must_env: std.native('must_env')",
		"md5: std.native('md5')",
		"file_stat: std.native('file_stat')",
	}

	for _, expected := range expectedFunctions {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateArmedLib result missing expected content: %s\nGot: %s", expected, result)
		}
	}
}
