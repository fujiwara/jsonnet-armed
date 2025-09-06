package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMD5Function(t *testing.T) {
	md5Func := HashFunctions[0].Func // md5 function

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "simple string",
			args:     []any{"hello"},
			expected: "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name:     "empty string",
			args:     []any{""},
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:        "non-string input",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := md5Func(tt.args)

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

func TestSHA256Function(t *testing.T) {
	sha256Func := HashFunctions[2].Func // sha256 function

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "simple string",
			args:     []any{"test"},
			expected: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name:     "empty string",
			args:     []any{""},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sha256Func(tt.args)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMD5FileFunction(t *testing.T) {
	md5FileFunc := HashFunctions[4].Func // md5_file function

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
			expected: "65a8e27d8879283831b664bd8b7f0ad4", // md5 of "Hello, World!"
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
			result, err := md5FileFunc(tt.args)

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
