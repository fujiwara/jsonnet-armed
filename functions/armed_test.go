package functions

import (
	"strings"
	"testing"
)

func TestGenerateArmedLib(t *testing.T) {
	funcs := GenerateAllFunctions(t.Context())
	result := GenerateArmedLib(funcs)

	// Verify it contains expected function definitions
	expectedFunctions := []string{
		"env: std.native('env')",
		"sha256: std.native('sha256')",
		"file_content: std.native('file_content')",
		"must_env: std.native('must_env')",
		"md5: std.native('md5')",
		"file_stat: std.native('file_stat')",
		"base64: std.native('base64')",
		"base64url: std.native('base64url')",
		"now: std.native('now')",
		"time_format: std.native('time_format')",
		"file_exists: std.native('file_exists')",
	}

	for _, expected := range expectedFunctions {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateArmedLib result missing expected content: %s\nGot: %s", expected, result)
		}
	}
}
