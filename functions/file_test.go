package functions_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fujiwara/jsonnet-armed"
)

func TestFileFunctions(t *testing.T) {
	ctx := context.Background()

	// Create test files with known content
	tmpDir := t.TempDir()
	
	// Create test file with simple content
	helloFile := filepath.Join(tmpDir, "hello.txt")
	helloContent := "Hello, World!"
	if err := os.WriteFile(helloFile, []byte(helloContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	// Create test file with JSON content
	jsonFile := filepath.Join(tmpDir, "test.json")
	jsonContent := `{"message": "test", "number": 42}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to create JSON test file: %v", err)
	}
	
	// Create test file with UTF-8 content
	utf8File := filepath.Join(tmpDir, "utf8.txt")
	utf8Content := "こんにちは、世界！"
	if err := os.WriteFile(utf8File, []byte(utf8Content), 0644); err != nil {
		t.Fatalf("failed to create UTF-8 test file: %v", err)
	}
	
	// Create empty file
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create empty test file: %v", err)
	}
	
	// Create a directory for testing
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "file_content with simple text file",
			jsonnet: fmt.Sprintf(`
			local file_content = std.native("file_content");
			{
				content: file_content("%s")
			}`, helloFile),
			expected: `{
				"content": "Hello, World!"
			}`,
		},
		{
			name: "file_content with JSON file",
			jsonnet: fmt.Sprintf(`
			local file_content = std.native("file_content");
			{
				raw_content: file_content("%s"),
				parsed: std.parseJson(file_content("%s"))
			}`, jsonFile, jsonFile),
			expected: `{
				"raw_content": "{\"message\": \"test\", \"number\": 42}",
				"parsed": {
					"message": "test",
					"number": 42
				}
			}`,
		},
		{
			name: "file_content with UTF-8 file",
			jsonnet: fmt.Sprintf(`
			local file_content = std.native("file_content");
			{
				content: file_content("%s")
			}`, utf8File),
			expected: `{
				"content": "こんにちは、世界！"
			}`,
		},
		{
			name: "file_content with empty file",
			jsonnet: fmt.Sprintf(`
			local file_content = std.native("file_content");
			{
				content: file_content("%s"),
				length: std.length(file_content("%s"))
			}`, emptyFile, emptyFile),
			expected: `{
				"content": "",
				"length": 0
			}`,
		},
		{
			name: "file_stat with regular file",
			jsonnet: fmt.Sprintf(`
			local file_stat = std.native("file_stat");
			local stat = file_stat("%s");
			{
				name: stat.name,
				size: stat.size,
				is_dir: stat.is_dir,
				has_mode: std.type(stat.mode) == "string",
				has_mod_time: std.type(stat.mod_time) == "number"
			}`, helloFile),
			expected: `{
				"name": "hello.txt",
				"size": 13,
				"is_dir": false,
				"has_mode": true,
				"has_mod_time": true
			}`,
		},
		{
			name: "file_stat with directory",
			jsonnet: fmt.Sprintf(`
			local file_stat = std.native("file_stat");
			local stat = file_stat("%s");
			{
				name: stat.name,
				is_dir: stat.is_dir,
				has_mode: std.type(stat.mode) == "string"
			}`, testDir),
			expected: `{
				"name": "testdir",
				"is_dir": true,
				"has_mode": true
			}`,
		},
		{
			name: "combining file_content and file_stat",
			jsonnet: fmt.Sprintf(`
			local file_content = std.native("file_content");
			local file_stat = std.native("file_stat");
			local content = file_content("%s");
			local stat = file_stat("%s");
			{
				content: content,
				content_length: std.length(content),
				file_size: stat.size,
				size_matches: std.length(content) == stat.size
			}`, helloFile, helloFile),
			expected: `{
				"content": "Hello, World!",
				"content_length": 13,
				"file_size": 13,
				"size_matches": true
			}`,
		},
		{
			name: "file_content with non-existent file",
			jsonnet: `
			local file_content = std.native("file_content");
			{
				content: file_content("/non/existent/file.txt")
			}`,
			expectError: true,
		},
		{
			name: "file_stat with non-existent file",
			jsonnet: `
			local file_stat = std.native("file_stat");
			{
				stat: file_stat("/non/existent/file.txt")
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
			}

			// Capture output
			var output bytes.Buffer
			armed.SetOutput(&output)
			defer armed.SetOutput(os.Stdout)

			// Run evaluation
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