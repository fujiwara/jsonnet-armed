package armed

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func TestArmedLibImport(t *testing.T) {
	// Set test environment variable
	t.Setenv("TEST_VAR", "test_value")

	// Create a temporary jsonnet file that imports armed.libsonnet
	jsonnetContent := `
local armed = import 'armed.libsonnet';

{
  sha256_test: armed.sha256('test'),
  env_test: armed.env('TEST_VAR', 'default_value'),
  env_default_test: armed.env('NONEXISTENT_VAR', 'default_value'),
  md5_test: armed.md5('hello'),
  all_functions: std.objectFields(armed),
}
`

	// Create a temporary file
	tmpfile := "/tmp/test_armed_import.jsonnet"
	err := os.WriteFile(tmpfile, []byte(jsonnetContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile)

	// Set up CLI with output capture
	var output bytes.Buffer
	cli := &CLI{
		Filename: tmpfile,
	}
	cli.SetWriter(&output)

	// Run jsonnet evaluation
	err = RunWithCLI(context.Background(), cli)
	if err != nil {
		t.Fatalf("RunWithCLI failed: %v", err)
	}

	result := output.String()

	// Verify expected content in output
	expectedContent := []string{
		"sha256_test",
		"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", // sha256('test')
		"env_test",
		"test_value", // From t.Setenv("TEST_VAR", "test_value")
		"env_default_test",
		"default_value", // From NONEXISTENT_VAR fallback
		"md5_test",
		"5d41402abc4b2a76b9719d911017c592", // md5('hello')
		"all_functions",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result, expected) {
			t.Errorf("Output missing expected content: %s\nGot: %s", expected, result)
		}
	}

	// Verify that all expected functions are listed
	expectedFunctions := []string{
		"env", "must_env",
		"md5", "sha1", "sha256", "sha512",
		"md5_file", "sha1_file", "sha256_file", "sha512_file",
		"file_content", "file_stat",
	}

	for _, funcName := range expectedFunctions {
		if !strings.Contains(result, `"`+funcName+`"`) {
			t.Errorf("Output missing function in all_functions: %s\nGot: %s", funcName, result)
		}
	}
}
