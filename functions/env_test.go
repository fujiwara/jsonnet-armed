package functions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fujiwara/jsonnet-armed"
	"github.com/google/go-cmp/cmp"
)

func TestEnvFunctions(t *testing.T) {
	ctx := context.Background()

	// Set test environment variables
	t.Setenv("TEST_ENV_VAR", "test-value")
	t.Setenv("TEST_NUMBER", "42")

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "env function with existing variable",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_ENV_VAR", "default")
			}`,
			expected: `{
				"value": "test-value"
			}`,
		},
		{
			name: "env function with non-existing variable returns default",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_UNSET_VAR", "default-value")
			}`,
			expected: `{
				"value": "default-value"
			}`,
		},
		{
			name: "env function with multiple calls",
			jsonnet: `
			local env = std.native("env");
			{
				existing: env("TEST_ENV_VAR", "default1"),
				missing: env("TEST_UNSET_VAR", "default2"),
				number: env("TEST_NUMBER", "0")
			}`,
			expected: `{
				"existing": "test-value",
				"missing": "default2",
				"number": "42"
			}`,
		},
		{
			name: "must_env with existing variable",
			jsonnet: `
			local must_env = std.native("must_env");
			{
				value: must_env("TEST_ENV_VAR")
			}`,
			expected: `{
				"value": "test-value"
			}`,
		},
		{
			name: "must_env with non-existing variable returns error",
			jsonnet: `
			local must_env = std.native("must_env");
			{
				value: must_env("TEST_UNSET_VAR")
			}`,
			expectError: true,
		},
		{
			name: "env with empty string as default",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_UNSET_VAR", "")
			}`,
			expected: `{
				"value": ""
			}`,
		},
		{
			name: "env with null as default",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_UNSET_VAR", null)
			}`,
			expected: `{
				"value": null
			}`,
		},
		{
			name: "env with object as default",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_UNSET_VAR", { key: "value" })
			}`,
			expected: `{
				"value": {
					"key": "value"
				}
			}`,
		},
		{
			name: "env with array as default",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_UNSET_VAR", [1, 2, 3])
			}`,
			expected: `{
				"value": [1, 2, 3]
			}`,
		},
		{
			name: "combining env functions with jsonnet logic",
			jsonnet: `
			local env = std.native("env");
			local envValue = env("TEST_ENV_VAR", "default");
			{
				original: envValue,
				uppercase: std.asciiUpper(envValue),
				length: std.length(envValue),
				conditional: if envValue == "test-value" then "matched" else "not matched"
			}`,
			expected: `{
				"original": "test-value",
				"uppercase": "TEST-VALUE",
				"length": 10,
				"conditional": "matched"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with jsonnet content
			tmpDir := t.TempDir()
			jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(jsonnetFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}

			// Create CLI config with output capture
			var output bytes.Buffer
			cli := &armed.CLI{
				Filename: jsonnetFile,
			}
			// Run evaluation
			cli.SetWriter(&output)
			err := armed.RunWithCLI(ctx, cli)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare JSON output
			compareJSON(t, output.String(), tt.expected)
		})
	}
}

func TestEnvFunctionsWithEmptyEnvVar(t *testing.T) {
	ctx := context.Background()

	// Set environment variable to empty string
	t.Setenv("TEST_EMPTY_VAR", "")

	tests := []struct {
		name     string
		jsonnet  string
		expected string
	}{
		{
			name: "env function with empty string environment variable",
			jsonnet: `
			local env = std.native("env");
			{
				value: env("TEST_EMPTY_VAR", "default")
			}`,
			expected: `{
				"value": "default"
			}`,
		},
		{
			name: "must_env with empty string environment variable",
			jsonnet: `
			local must_env = std.native("must_env");
			{
				value: must_env("TEST_EMPTY_VAR")
			}`,
			expected: `{
				"value": ""
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(jsonnetFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}

			var output bytes.Buffer
			cli := &armed.CLI{
				Filename: jsonnetFile,
			}
			cli.SetWriter(&output)

			if err := armed.RunWithCLI(ctx, cli); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			compareJSON(t, output.String(), tt.expected)
		})
	}
}

// compareJSON compares two JSON strings for equality
func compareJSON(t *testing.T, got, want string) {
	t.Helper()

	var gotJSON, wantJSON interface{}

	if err := json.Unmarshal([]byte(got), &gotJSON); err != nil {
		t.Fatalf("failed to parse got JSON: %v\nJSON: %s", err, got)
	}

	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatalf("failed to parse want JSON: %v\nJSON: %s", err, want)
	}

	if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}
