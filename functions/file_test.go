package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFileContentFunction(t *testing.T) {
	fileContentFunc := FileFunctions[0].Func // file_content function

	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "existing file",
			args:     []any{testFile},
			expected: "Hello, World!",
		},
		{
			name:        "non-existent file",
			args:        []any{"/non/existent/file.txt"},
			expectError: true,
		},
		{
			name:        "non-string filename",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fileContentFunc(tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFileStatFunction(t *testing.T) {
	fileStatFunc := FileFunctions[1].Func // file_stat function

	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expectError bool
	}{
		{
			name: "existing file",
			args: []any{testFile},
		},
		{
			name:        "non-existent file",
			args:        []any{"/non/existent/file.txt"},
			expectError: true,
		},
		{
			name:        "non-string filename",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fileStatFunc(tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that result is a map with expected keys
			stat, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T", result)
			}

			expectedKeys := []string{"name", "size", "mode", "mod_time", "is_dir"}
			for _, key := range expectedKeys {
				if _, exists := stat[key]; !exists {
					t.Errorf("missing key %s in stat result", key)
				}
			}

			// Verify specific values
			if stat["name"] != "test.txt" {
				t.Errorf("expected name 'test.txt', got %v", stat["name"])
			}
			if stat["size"] != int64(13) { // "Hello, World!" is 13 bytes
				t.Errorf("expected size 13, got %v", stat["size"])
			}
			if stat["is_dir"] != false {
				t.Errorf("expected is_dir false, got %v", stat["is_dir"])
			}
		})
	}
}

func TestFileExistsFunction(t *testing.T) {
	fileExistsFunc := FileFunctions[2].Func // file_exists function

	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    bool
		expectError bool
	}{
		{
			name:     "existing file",
			args:     []any{testFile},
			expected: true,
		},
		{
			name:     "non-existent file",
			args:     []any{"/non/existent/file.txt"},
			expected: false,
		},
		{
			name:        "non-string filename",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fileExistsFunc(tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
