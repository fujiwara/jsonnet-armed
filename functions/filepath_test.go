package functions_test

import (
	"testing"
)

func TestBasename(t *testing.T) {
	fn, err := getPathFunction("basename")
	if err != nil {
		t.Fatalf("failed to get basename function: %v", err)
	}

	tests := []struct {
		name     string
		args     []any
		expected any
		wantErr  bool
	}{
		{name: "absolute path", args: []any{"/usr/local/bin/program"}, expected: "program"},
		{name: "relative path", args: []any{"dir/file.txt"}, expected: "file.txt"},
		{name: "filename only", args: []any{"file.txt"}, expected: "file.txt"},
		{name: "trailing slash", args: []any{"/usr/local/"}, expected: "local"},
		{name: "root", args: []any{"/"}, expected: "/"},
		{name: "dotfile", args: []any{"/home/user/.bashrc"}, expected: ".bashrc"},
		{name: "empty string", args: []any{"."}, expected: "."},
		{name: "non-string arg", args: []any{123}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDirname(t *testing.T) {
	fn, err := getPathFunction("dirname")
	if err != nil {
		t.Fatalf("failed to get dirname function: %v", err)
	}

	tests := []struct {
		name     string
		args     []any
		expected any
		wantErr  bool
	}{
		{name: "absolute path", args: []any{"/usr/local/bin/program"}, expected: "/usr/local/bin"},
		{name: "relative path", args: []any{"dir/file.txt"}, expected: "dir"},
		{name: "filename only", args: []any{"file.txt"}, expected: "."},
		{name: "root file", args: []any{"/file.txt"}, expected: "/"},
		{name: "nested path", args: []any{"/a/b/c/d"}, expected: "/a/b/c"},
		{name: "dotfile", args: []any{"/home/user/.bashrc"}, expected: "/home/user"},
		{name: "non-string arg", args: []any{true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtname(t *testing.T) {
	fn, err := getPathFunction("extname")
	if err != nil {
		t.Fatalf("failed to get extname function: %v", err)
	}

	tests := []struct {
		name     string
		args     []any
		expected any
		wantErr  bool
	}{
		{name: "simple extension", args: []any{"file.txt"}, expected: ".txt"},
		{name: "double extension", args: []any{"archive.tar.gz"}, expected: ".gz"},
		{name: "no extension", args: []any{"Makefile"}, expected: ""},
		{name: "dotfile", args: []any{".bashrc"}, expected: ".bashrc"},
		{name: "path with extension", args: []any{"/usr/local/file.json"}, expected: ".json"},
		{name: "dotfile with extension", args: []any{".config.yaml"}, expected: ".yaml"},
		{name: "non-string arg", args: []any{nil}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPathJoin(t *testing.T) {
	fn, err := getPathFunction("path_join")
	if err != nil {
		t.Fatalf("failed to get path_join function: %v", err)
	}

	tests := []struct {
		name     string
		args     []any
		expected any
		wantErr  bool
	}{
		{name: "two elements", args: []any{[]any{"usr", "local"}}, expected: "usr/local"},
		{name: "absolute path", args: []any{[]any{"/usr", "local", "bin"}}, expected: "/usr/local/bin"},
		{name: "with file", args: []any{[]any{"/home", "user", "file.txt"}}, expected: "/home/user/file.txt"},
		{name: "single element", args: []any{[]any{"file.txt"}}, expected: "file.txt"},
		{name: "empty array", args: []any{[]any{}}, expected: ""},
		{name: "with dot", args: []any{[]any{".", "src", "main.go"}}, expected: "src/main.go"},
		{name: "non-array arg", args: []any{"not-an-array"}, wantErr: true},
		{name: "non-string element", args: []any{[]any{"valid", 123}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
