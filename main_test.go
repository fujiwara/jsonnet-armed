package armed_test

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

func TestRunWithCLI(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		extStr      map[string]string
		extCode     map[string]string
		expected    string
		expectError bool
	}{
		{
			name: "basic evaluation",
			jsonnet: `{
				foo: "bar",
				baz: 123
			}`,
			expected: `{
				"foo": "bar",
				"baz": 123
			}`,
		},
		{
			name: "with external string variables",
			jsonnet: `{
				env: std.extVar("env"),
				region: std.extVar("region")
			}`,
			extStr: map[string]string{
				"env":    "production",
				"region": "us-west-2",
			},
			expected: `{
				"env": "production",
				"region": "us-west-2"
			}`,
		},
		{
			name: "with external code variables",
			jsonnet: `{
				replicas: std.extVar("replicas"),
				debug: std.extVar("debug")
			}`,
			extCode: map[string]string{
				"replicas": "3",
				"debug":    "true",
			},
			expected: `{
				"replicas": 3,
				"debug": true
			}`,
		},
		{
			name: "mixed external variables",
			jsonnet: `{
				name: std.extVar("name"),
				count: std.extVar("count"),
				config: std.extVar("config")
			}`,
			extStr: map[string]string{
				"name": "test-app",
			},
			extCode: map[string]string{
				"count":  "5 * 2",
				"config": `{ enabled: true, timeout: 30 }`,
			},
			expected: `{
				"name": "test-app",
				"count": 10,
				"config": {
					"enabled": true,
					"timeout": 30
				}
			}`,
		},
		{
			name: "complex jsonnet with functions",
			jsonnet: `
			local multiply(x, y) = x * y;
			{
				simple: "value",
				computed: multiply(3, 7),
				array: [1, 2, 3],
				nested: {
					key: "nested value",
					sum: std.foldl(function(x, y) x + y, [1, 2, 3, 4, 5], 0)
				}
			}`,
			expected: `{
				"simple": "value",
				"computed": 21,
				"array": [1, 2, 3],
				"nested": {
					"key": "nested value",
					"sum": 15
				}
			}`,
		},
		{
			name: "jsonnet with conditionals",
			jsonnet: `{
				env: std.extVar("env"),
				replicas: if std.extVar("env") == "production" then 3 else 1,
				debug: std.extVar("env") != "production"
			}`,
			extStr: map[string]string{
				"env": "production",
			},
			expected: `{
				"env": "production",
				"replicas": 3,
				"debug": false
			}`,
		},
		{
			name: "error: missing external variable",
			jsonnet: `{
				value: std.extVar("missing")
			}`,
			expectError: true,
		},
		{
			name: "error: invalid jsonnet syntax",
			jsonnet: `{
				invalid: syntax error
			}`,
			expectError: true,
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

			// Create CLI config
			cli := &armed.CLI{
				Filename: jsonnetFile,
				ExtStr:   tt.extStr,
				ExtCode:  tt.extCode,
			}

			// Capture output
			var output bytes.Buffer
			armed.SetOutput(&output)
			defer armed.SetOutput(os.Stdout) // Restore default output

			// Run evaluation
			err := armed.RunWithCLI(ctx, cli)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return // Skip further checks for error cases
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare JSON output
			compareJSON(t, output.String(), tt.expected)
		})
	}
}

func TestRunWithCLIOutputToFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		jsonnet  string
		extStr   map[string]string
		extCode  map[string]string
		expected string
	}{
		{
			name: "output to file",
			jsonnet: `{
				app: std.extVar("app"),
				version: std.extVar("version")
			}`,
			extStr: map[string]string{
				"app": "my-app",
			},
			extCode: map[string]string{
				"version": `"1.2.3"`,
			},
			expected: `{
				"app": "my-app",
				"version": "1.2.3"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp jsonnet file
			jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(jsonnetFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}

			// Create output file path
			outputFile := filepath.Join(tmpDir, "output.json")

			// Create CLI config
			cli := &armed.CLI{
				Filename:   jsonnetFile,
				OutputFile: outputFile,
				ExtStr:     tt.extStr,
				ExtCode:    tt.extCode,
			}

			// Run evaluation
			if err := armed.RunWithCLI(ctx, cli); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read output file
			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			// Compare JSON output
			compareJSON(t, string(output), tt.expected)
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
