package functions_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fujiwara/jsonnet-armed"
)

func TestHashFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "sha256 with simple string",
			jsonnet: `
			local sha256 = std.native("sha256");
			{
				hash: sha256("hello")
			}`,
			expected: `{
				"hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
			}`,
		},
		{
			name: "sha256 with empty string",
			jsonnet: `
			local sha256 = std.native("sha256");
			{
				hash: sha256("")
			}`,
			expected: `{
				"hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}`,
		},
		{
			name: "sha256 with multiple calls",
			jsonnet: `
			local sha256 = std.native("sha256");
			{
				hello: sha256("hello"),
				world: sha256("world"),
				hello_world: sha256("helloworld")
			}`,
			expected: `{
				"hello": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"world": "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7",
				"hello_world": "936a185caaa266bb9cbe981e9e05cb78cd732b0b3280eb944412bb6f8f8f07af"
			}`,
		},
		{
			name: "sha256 with UTF-8 string",
			jsonnet: `
			local sha256 = std.native("sha256");
			{
				hash: sha256("こんにちは")
			}`,
			expected: `{
				"hash": "125aeadf27b0459b8760c13a3d80912dfa8a81a68261906f60d87f4a0268646c"
			}`,
		},
		{
			name: "sha256 combining with jsonnet logic",
			jsonnet: `
			local sha256 = std.native("sha256");
			local data = "test-data";
			local hash = sha256(data);
			{
				original: data,
				hash: hash,
				prefix: std.substr(hash, 0, 8),
				length: std.length(hash)
			}`,
			expected: `{
				"original": "test-data",
				"hash": "a186000422feab857329c684e9fe91412b1a5db084100b37a98cfc95b62aa867",
				"prefix": "a1860004",
				"length": 64
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