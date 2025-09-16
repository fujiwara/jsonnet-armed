package armed_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
			err := cli.Run(ctx)

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
			if err := cli.Run(ctx); err != nil {
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
			err = cli.Run(ctx)

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
			err := cli.Run(ctx)

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
			err = cli.Run(ctx)

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
			err := cli.Run(ctx)

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

// TestAtomicFileWrite tests that file writing is atomic
// TestWriteIfChanged tests the --write-if-changed functionality
func TestWriteIfChanged(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	t.Run("write new file", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "new.jsonnet")
		outputFile := filepath.Join(tmpDir, "new_output.json")

		jsonnet := `{"test": "new file"}`
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		cli := &armed.CLI{
			Filename:       jsonnetFile,
			OutputFile:     outputFile,
			WriteIfChanged: true,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// File should exist
		if _, err := os.Stat(outputFile); err != nil {
			t.Errorf("output file should exist: %v", err)
		}
	})

	t.Run("skip unchanged file", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "unchanged.jsonnet")
		outputFile := filepath.Join(tmpDir, "unchanged_output.json")

		jsonnet := `{"test": "unchanged"}`
		// Use the exact format that jsonnet produces (with newline)
		expectedOutput := `{
   "test": "unchanged"
}
`

		// Write jsonnet file
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		// Write existing output file
		if err := os.WriteFile(outputFile, []byte(expectedOutput), 0644); err != nil {
			t.Fatalf("failed to write output file: %v", err)
		}

		// Get original modification time
		originalInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}
		originalModTime := originalInfo.ModTime()

		// Sleep to ensure time difference
		time.Sleep(10 * time.Millisecond)

		cli := &armed.CLI{
			Filename:       jsonnetFile,
			OutputFile:     outputFile,
			WriteIfChanged: true,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check modification time hasn't changed
		newInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}

		if !newInfo.ModTime().Equal(originalModTime) {
			t.Errorf("file should not have been modified: original=%v, new=%v",
				originalModTime, newInfo.ModTime())
		}
	})

	t.Run("update changed file", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "changed.jsonnet")
		outputFile := filepath.Join(tmpDir, "changed_output.json")

		jsonnet := `{"test": "changed", "value": 42}`
		oldOutput := `{"test":"old","value":1}`

		// Write jsonnet file
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		// Write existing output file with different content
		if err := os.WriteFile(outputFile, []byte(oldOutput), 0644); err != nil {
			t.Fatalf("failed to write output file: %v", err)
		}

		// Get original modification time
		originalInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}
		originalModTime := originalInfo.ModTime()

		// Sleep to ensure time difference
		time.Sleep(10 * time.Millisecond)

		cli := &armed.CLI{
			Filename:       jsonnetFile,
			OutputFile:     outputFile,
			WriteIfChanged: true,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check modification time has changed
		newInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}

		if newInfo.ModTime().Equal(originalModTime) {
			t.Error("file should have been modified")
		}

		// Verify new content
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(content, &result); err != nil {
			t.Errorf("invalid JSON: %v", err)
		}

		if result["test"] != "changed" || result["value"] != float64(42) {
			t.Errorf("unexpected content: %v", result)
		}
	})

	t.Run("update file with different size", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "size.jsonnet")
		outputFile := filepath.Join(tmpDir, "size_output.json")

		jsonnet := `{"test": "new content with different size"}`
		oldOutput := `{"test":"old"}`

		// Write jsonnet file
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		// Write existing output file with different size
		if err := os.WriteFile(outputFile, []byte(oldOutput), 0644); err != nil {
			t.Fatalf("failed to write output file: %v", err)
		}

		cli := &armed.CLI{
			Filename:       jsonnetFile,
			OutputFile:     outputFile,
			WriteIfChanged: true,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify new content
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(content, &result); err != nil {
			t.Errorf("invalid JSON: %v", err)
		}

		if result["test"] != "new content with different size" {
			t.Errorf("unexpected content: %v", result)
		}
	})

	t.Run("disabled by default", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "default.jsonnet")
		outputFile := filepath.Join(tmpDir, "default_output.json")

		jsonnet := `{"test": "default"}`
		// Use the exact format that jsonnet produces (with newline)
		expectedOutput := `{
   "test": "default"
}
`

		// Write jsonnet file
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		// Write existing output file
		if err := os.WriteFile(outputFile, []byte(expectedOutput), 0644); err != nil {
			t.Fatalf("failed to write output file: %v", err)
		}

		// Get original modification time
		originalInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}
		originalModTime := originalInfo.ModTime()

		// Sleep to ensure time difference
		time.Sleep(10 * time.Millisecond)

		cli := &armed.CLI{
			Filename:       jsonnetFile,
			OutputFile:     outputFile,
			WriteIfChanged: false, // Explicitly disabled (default)
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check modification time HAS changed (always writes when disabled)
		newInfo, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}

		if newInfo.ModTime().Equal(originalModTime) {
			t.Error("file should have been modified when WriteIfChanged is disabled")
		}
	})
}

func TestAtomicFileWrite(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Test concurrent writes to ensure atomicity
	t.Run("concurrent writes", func(t *testing.T) {
		outputFile := filepath.Join(tmpDir, "concurrent_output.json")

		// Create different jsonnet files for each writer to avoid race condition
		jsonnetFiles := make([]string, 3)
		contents := []string{
			`{"writer": 1, "data": "first"}`,
			`{"writer": 2, "data": "second"}`,
			`{"writer": 3, "data": "third"}`,
		}

		for i := range jsonnetFiles {
			jsonnetFiles[i] = filepath.Join(tmpDir, fmt.Sprintf("test_%d.jsonnet", i))
			if err := os.WriteFile(jsonnetFiles[i], []byte(contents[i]), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}
		}

		// Run concurrent writes to the same output file
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				cli := &armed.CLI{
					Filename:   jsonnetFiles[index],
					OutputFile: outputFile,
				}

				if err := cli.Run(ctx); err != nil {
					t.Logf("writer %d error: %v", index, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify the output file exists and is valid JSON
		data, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("output file contains invalid JSON: %v\nContent: %s", err, string(data))
		}

		// The file should contain complete JSON from one of the writers
		if writer, ok := result["writer"].(float64); ok {
			if writer < 1 || writer > 3 {
				t.Errorf("unexpected writer value: %v", writer)
			}
		} else {
			t.Errorf("writer field not found or not a number")
		}
	})

	// Test that intermediate temporary files are cleaned up
	t.Run("temp file cleanup", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "test_cleanup.jsonnet")
		outputFile := filepath.Join(tmpDir, "cleanup_output.json")

		jsonnet := `{"test": "cleanup"}`
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		cli := &armed.CLI{
			Filename:   jsonnetFile,
			OutputFile: outputFile,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that no .tmp files remain
		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".tmp") {
				t.Errorf("temporary file not cleaned up: %s", entry.Name())
			}
		}
	})

	// Test that file permissions are preserved
	t.Run("preserve permissions", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "test_perms.jsonnet")
		outputFile := filepath.Join(tmpDir, "perms_output.json")

		jsonnet := `{"test": "permissions"}`
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		cli := &armed.CLI{
			Filename:   jsonnetFile,
			OutputFile: outputFile,
		}

		if err := cli.Run(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check file permissions
		info, err := os.Stat(outputFile)
		if err != nil {
			t.Fatalf("failed to stat output file: %v", err)
		}

		// Should have 0644 permissions
		expectedMode := os.FileMode(0644)
		if info.Mode().Perm() != expectedMode {
			t.Errorf("unexpected file permissions: got %v, want %v", info.Mode().Perm(), expectedMode)
		}
	})

	// Test error handling during write
	t.Run("write error handling", func(t *testing.T) {
		jsonnetFile := filepath.Join(tmpDir, "test_error.jsonnet")
		// Try to write to a directory that doesn't exist
		outputFile := filepath.Join(tmpDir, "nonexistent", "subdir", "output.json")

		jsonnet := `{"test": "error"}`
		if err := os.WriteFile(jsonnetFile, []byte(jsonnet), 0644); err != nil {
			t.Fatalf("failed to write jsonnet file: %v", err)
		}

		cli := &armed.CLI{
			Filename:   jsonnetFile,
			OutputFile: outputFile,
		}

		err := cli.Run(ctx)
		if err == nil {
			t.Fatal("expected error for non-existent directory, got nil")
		}

		// The error should mention the directory issue
		if !strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
