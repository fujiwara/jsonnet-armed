package functions_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fujiwara/jsonnet-armed"
)

func TestBase64Functions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "base64 encode simple string",
			jsonnet: `
			local base64 = std.native("base64");
			base64("Hello, World!")`,
			expected: `"SGVsbG8sIFdvcmxkIQ=="`,
		},
		{
			name: "base64 encode empty string",
			jsonnet: `
			local base64 = std.native("base64");
			base64("")`,
			expected: `""`,
		},
		{
			name: "base64 encode special characters",
			jsonnet: `
			local base64 = std.native("base64");
			base64("The quick brown fox jumps over the lazy dog")`,
			expected: `"VGhlIHF1aWNrIGJyb3duIGZveCBqdW1wcyBvdmVyIHRoZSBsYXp5IGRvZw=="`,
		},
		{
			name: "base64url encode simple string",
			jsonnet: `
			local base64url = std.native("base64url");
			base64url("Hello, World!")`,
			expected: `"SGVsbG8sIFdvcmxkIQ=="`,
		},
		{
			name: "base64url encode with URL-unsafe characters",
			jsonnet: `
			local base64url = std.native("base64url");
			base64url("??>>")`,
			expected: `"Pz8-Pg=="`,
		},
		{
			name: "base64 vs base64url difference",
			jsonnet: `
			local base64 = std.native("base64");
			local base64url = std.native("base64url");
			{
				standard: base64("??>>"),
				urlsafe: base64url("??>>")
			}`,
			expected: `{
				"standard": "Pz8+Pg==",
				"urlsafe": "Pz8-Pg=="
			}`,
		},
		{
			name: "base64 with non-string input",
			jsonnet: `
			local base64 = std.native("base64");
			base64(123)`,
			expectError: true,
		},
		{
			name: "base64url with non-string input",
			jsonnet: `
			local base64url = std.native("base64url");
			base64url(123)`,
			expectError: true,
		},
		{
			name: "base64 encode unicode string",
			jsonnet: `
			local base64 = std.native("base64");
			base64("こんにちは世界")`,
			expected: `"44GT44KT44Gr44Gh44Gv5LiW55WM"`,
		},
		{
			name: "base64 functions with jsonnet logic",
			jsonnet: `
			local base64 = std.native("base64");
			local data = "test-data";
			local encoded = base64(data);
			{
				original: data,
				encoded: encoded,
				length: std.length(encoded)
			}`,
			expected: `{
				"original": "test-data",
				"encoded": "dGVzdC1kYXRh",
				"length": 12
			}`,
		},
		{
			name: "armed library import with base64",
			jsonnet: `
			local armed = import 'armed.libsonnet';
			{
				base64: armed.base64("hello"),
				base64url: armed.base64url("hello")
			}`,
			expected: `{
				"base64": "aGVsbG8=",
				"base64url": "aGVsbG8="
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
