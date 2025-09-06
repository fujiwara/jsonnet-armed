package armed_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed"
	"github.com/google/go-cmp/cmp"
)

// slowWriter simulates a slow output writer that blocks during Write
type slowWriter struct {
	delay time.Duration
}

func (sw *slowWriter) Write(p []byte) (n int, err error) {
	// Sleep longer than timeout to simulate slow output
	time.Sleep(sw.delay)
	return len(p), nil
}

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

			// Capture output
			var output bytes.Buffer

			// Create CLI config
			cli := &armed.CLI{
				Filename: jsonnetFile,
				ExtStr:   tt.extStr,
				ExtCode:  tt.extCode,
			}
			cli.SetWriter(&output)

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

func TestRunWithCLIFromStdin(t *testing.T) {
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
			name: "basic stdin evaluation",
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
			name: "stdin with external string variables",
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
			name: "stdin with external code variables",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin
			originalStdin := os.Stdin
			defer func() { os.Stdin = originalStdin }()

			// Create pipe for stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdin = r

			// Write jsonnet content to pipe
			go func() {
				defer w.Close()
				io.WriteString(w, tt.jsonnet)
			}()

			// Capture output
			var output bytes.Buffer

			// Create CLI config with "-" as filename
			cli := &armed.CLI{
				Filename: "-",
				ExtStr:   tt.extStr,
				ExtCode:  tt.extCode,
			}
			cli.SetWriter(&output)

			// Run evaluation
			err = armed.RunWithCLI(ctx, cli)

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

func TestRunWithCLITimeout(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		timeout     time.Duration
		expectError bool
		errorMatch  string
	}{
		{
			name: "normal evaluation without timeout",
			jsonnet: `{
				message: "hello",
				number: 42
			}`,
			timeout:     0, // no timeout
			expectError: false,
		},
		{
			name: "normal evaluation with sufficient timeout",
			jsonnet: `{
				message: "hello",
				timestamp: std.native("now")()
			}`,
			timeout:     5 * time.Second, // plenty of time
			expectError: false,
		},
		{
			name: "evaluation with short timeout should timeout",
			jsonnet: `{
				// Very heavy nested computation that should definitely timeout
				message: "very heavy computation",
				data: [
					[x + y + z + w for w in std.range(1, 20)]
					for x in std.range(1, 30) 
					for y in std.range(1, 30) 
					for z in std.range(1, 30)
				],
				more_data: std.foldl(function(acc, x) 
					std.foldl(function(inner_acc, y) inner_acc + (x * y), std.range(1, 100), acc), 
					std.range(1, 100), 0)
			}`,
			timeout:     50 * time.Millisecond, // very short timeout
			expectError: true,
			errorMatch:  "evaluation timed out after",
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

			// Capture output
			var output bytes.Buffer

			// Create CLI config with timeout
			cli := &armed.CLI{
				Filename: jsonnetFile,
				Timeout:  tt.timeout,
			}
			cli.SetWriter(&output)

			// Run evaluation
			err := armed.RunWithCLI(ctx, cli)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errorMatch != "" && !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorMatch, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Should have valid JSON output
				var result map[string]interface{}
				if err := json.Unmarshal(output.Bytes(), &result); err != nil {
					t.Errorf("output is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestRunWithCLITimeoutFromStdin(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		timeout     time.Duration
		expectError bool
		errorMatch  string
		delayWrite  bool
	}{
		{
			name: "normal stdin evaluation with sufficient timeout",
			jsonnet: `{
				message: "hello from stdin",
				timestamp: std.native("now")()
			}`,
			timeout:     5 * time.Second,
			expectError: false,
			delayWrite:  false,
		},
		{
			name:        "stdin read should timeout when no data available",
			jsonnet:     "", // This will never be used since stdin read will timeout
			timeout:     100 * time.Millisecond,
			expectError: true,
			errorMatch:  "evaluation timed out after",
			delayWrite:  true, // Special flag to simulate slow/no stdin data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var output bytes.Buffer

			// Create CLI config with timeout and stdin
			cli := &armed.CLI{
				Filename: "-",
				Timeout:  tt.timeout,
			}
			cli.SetWriter(&output)

			// Mock stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			defer r.Close()
			defer w.Close()

			os.Stdin = r

			// Write to stdin in goroutine
			go func() {
				defer w.Close()
				if tt.delayWrite {
					// Simulate slow stdin by waiting longer than timeout
					time.Sleep(tt.timeout * 2)
				}
				if tt.jsonnet != "" {
					w.Write([]byte(tt.jsonnet))
				}
			}()

			// Run evaluation
			err = armed.RunWithCLI(ctx, cli)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errorMatch != "" && !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorMatch, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should have valid JSON output
			var result map[string]interface{}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Errorf("output is not valid JSON: %v", err)
			}
		})
	}
}

func TestRunWithCLITimeoutSlowOutput(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		timeout     time.Duration
		expectError bool
		errorMatch  string
		slowOutput  bool
	}{
		{
			name: "normal output with sufficient timeout",
			jsonnet: `{
				message: "normal output",
				number: 123
			}`,
			timeout:     5 * time.Second,
			expectError: false,
			slowOutput:  false,
		},
		{
			name: "simple jsonnet but output write should timeout",
			jsonnet: `{
				message: "simple output that will timeout during write",
				number: 42
			}`,
			timeout:     50 * time.Millisecond,
			expectError: true,
			errorMatch:  "evaluation timed out after",
			slowOutput:  true,
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

			// Set up output writer
			var output io.Writer
			if tt.slowOutput {
				output = &slowWriter{delay: tt.timeout * 2} // Slower than timeout
			} else {
				output = &bytes.Buffer{}
			}

			// Create CLI config
			cli := &armed.CLI{
				Filename: jsonnetFile,
				Timeout:  tt.timeout,
			}
			cli.SetWriter(output)

			// Run evaluation
			err := armed.RunWithCLI(ctx, cli)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errorMatch != "" && !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorMatch, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// For successful cases, verify output if it's a buffer
			if buf, ok := output.(*bytes.Buffer); ok {
				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("output is not valid JSON: %v", err)
				}
			}
		})
	}
}
