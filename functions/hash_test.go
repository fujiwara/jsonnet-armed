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

func TestHashFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "md5 with simple string",
			jsonnet: `
			local md5 = std.native("md5");
			{
				hash: md5("hello")
			}`,
			expected: `{
				"hash": "5d41402abc4b2a76b9719d911017c592"
			}`,
		},
		{
			name: "md5 with empty string",
			jsonnet: `
			local md5 = std.native("md5");
			{
				hash: md5("")
			}`,
			expected: `{
				"hash": "d41d8cd98f00b204e9800998ecf8427e"
			}`,
		},
		{
			name: "sha1 with simple string",
			jsonnet: `
			local sha1 = std.native("sha1");
			{
				hash: sha1("hello")
			}`,
			expected: `{
				"hash": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
			}`,
		},
		{
			name: "sha1 with empty string",
			jsonnet: `
			local sha1 = std.native("sha1");
			{
				hash: sha1("")
			}`,
			expected: `{
				"hash": "da39a3ee5e6b4b0d3255bfef95601890afd80709"
			}`,
		},
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
			name: "sha512 with simple string",
			jsonnet: `
			local sha512 = std.native("sha512");
			{
				hash: sha512("hello")
			}`,
			expected: `{
				"hash": "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
			}`,
		},
		{
			name: "sha512 with empty string",
			jsonnet: `
			local sha512 = std.native("sha512");
			{
				hash: sha512("")
			}`,
			expected: `{
				"hash": "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
			}`,
		},
		{
			name: "multiple hash functions with same input",
			jsonnet: `
			local md5 = std.native("md5");
			local sha1 = std.native("sha1");
			local sha256 = std.native("sha256");
			local sha512 = std.native("sha512");
			{
				md5: md5("hello"),
				sha1: sha1("hello"),
				sha256: sha256("hello"),
				sha512: sha512("hello")
			}`,
			expected: `{
				"md5": "5d41402abc4b2a76b9719d911017c592",
				"sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
				"sha256": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"sha512": "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
			}`,
		},
		{
			name: "hash functions with UTF-8 string",
			jsonnet: `
			local md5 = std.native("md5");
			local sha1 = std.native("sha1");
			local sha256 = std.native("sha256");
			local sha512 = std.native("sha512");
			{
				md5: md5("こんにちは"),
				sha1: sha1("こんにちは"),
				sha256: sha256("こんにちは"),
				sha512: sha512("こんにちは")
			}`,
			expected: `{
				"md5": "c0e89a293bd36c7a768e4e9d2c5475a8",
				"sha1": "20427a708c3f6f07cf12ab23557982d9e6d23b61",
				"sha256": "125aeadf27b0459b8760c13a3d80912dfa8a81a68261906f60d87f4a0268646c",
				"sha512": "bb2b0b573e976d4240fd775e3b0d8c8fcbd058d832fe451214db9d604dc7b3817f0b1b030d27488c96fc0e008228172acdd5e15c26f6543d5f48dc75d8d9a662"
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

func TestHashFileFunctions(t *testing.T) {
	ctx := context.Background()

	// Create test files with known content
	tmpDir := t.TempDir()
	
	// Create test file with "hello" content
	helloFile := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(helloFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	// Create test file with empty content
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create empty test file: %v", err)
	}
	
	// Create test file with UTF-8 content
	utf8File := filepath.Join(tmpDir, "utf8.txt")
	if err := os.WriteFile(utf8File, []byte("こんにちは"), 0644); err != nil {
		t.Fatalf("failed to create utf8 test file: %v", err)
	}

	tests := []struct {
		name        string
		jsonnet     string
		expected    string
		expectError bool
	}{
		{
			name: "md5_file with hello file",
			jsonnet: fmt.Sprintf(`
			local md5_file = std.native("md5_file");
			{
				hash: md5_file("%s")
			}`, helloFile),
			expected: `{
				"hash": "5d41402abc4b2a76b9719d911017c592"
			}`,
		},
		{
			name: "sha256_file with hello file",
			jsonnet: fmt.Sprintf(`
			local sha256_file = std.native("sha256_file");
			{
				hash: sha256_file("%s")
			}`, helloFile),
			expected: `{
				"hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
			}`,
		},
		{
			name: "sha256_file with empty file",
			jsonnet: fmt.Sprintf(`
			local sha256_file = std.native("sha256_file");
			{
				hash: sha256_file("%s")
			}`, emptyFile),
			expected: `{
				"hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}`,
		},
		{
			name: "multiple file hash functions with same file",
			jsonnet: fmt.Sprintf(`
			local md5_file = std.native("md5_file");
			local sha1_file = std.native("sha1_file");
			local sha256_file = std.native("sha256_file");
			local sha512_file = std.native("sha512_file");
			{
				md5: md5_file("%s"),
				sha1: sha1_file("%s"),
				sha256: sha256_file("%s"),
				sha512: sha512_file("%s")
			}`, helloFile, helloFile, helloFile, helloFile),
			expected: `{
				"md5": "5d41402abc4b2a76b9719d911017c592",
				"sha1": "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
				"sha256": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"sha512": "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
			}`,
		},
		{
			name: "file hash functions with UTF-8 file",
			jsonnet: fmt.Sprintf(`
			local md5_file = std.native("md5_file");
			local sha1_file = std.native("sha1_file");
			local sha256_file = std.native("sha256_file");
			local sha512_file = std.native("sha512_file");
			{
				md5: md5_file("%s"),
				sha1: sha1_file("%s"),
				sha256: sha256_file("%s"),
				sha512: sha512_file("%s")
			}`, utf8File, utf8File, utf8File, utf8File),
			expected: `{
				"md5": "c0e89a293bd36c7a768e4e9d2c5475a8",
				"sha1": "20427a708c3f6f07cf12ab23557982d9e6d23b61",
				"sha256": "125aeadf27b0459b8760c13a3d80912dfa8a81a68261906f60d87f4a0268646c",
				"sha512": "bb2b0b573e976d4240fd775e3b0d8c8fcbd058d832fe451214db9d604dc7b3817f0b1b030d27488c96fc0e008228172acdd5e15c26f6543d5f48dc75d8d9a662"
			}`,
		},
		{
			name: "sha256_file with non-existent file",
			jsonnet: `
			local sha256_file = std.native("sha256_file");
			{
				hash: sha256_file("/non/existent/file.txt")
			}`,
			expectError: true,
		},
		{
			name: "comparing string and file hash results",
			jsonnet: fmt.Sprintf(`
			local sha256 = std.native("sha256");
			local sha256_file = std.native("sha256_file");
			{
				string_hash: sha256("hello"),
				file_hash: sha256_file("%s"),
				matches: sha256("hello") == sha256_file("%s")
			}`, helloFile, helloFile),
			expected: `{
				"string_hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"file_hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"matches": true
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