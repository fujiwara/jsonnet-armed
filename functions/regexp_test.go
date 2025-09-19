package functions_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRegexMatchFunction(t *testing.T) {
	regexMatchFunc, err := getRegexpFunction("regex_match")
	if err != nil {
		t.Fatalf("failed to get regex_match function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    bool
		expectError bool
	}{
		{
			name:     "simple match",
			args:     []any{"hello", "hello world"},
			expected: true,
		},
		{
			name:     "no match",
			args:     []any{"xyz", "hello world"},
			expected: false,
		},
		{
			name:     "regex pattern match",
			args:     []any{"^[a-z]+$", "hello"},
			expected: true,
		},
		{
			name:     "regex pattern no match",
			args:     []any{"^[0-9]+$", "hello"},
			expected: false,
		},
		{
			name:     "email validation match",
			args:     []any{"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "test@example.com"},
			expected: true,
		},
		{
			name:     "email validation no match",
			args:     []any{"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "invalid-email"},
			expected: false,
		},
		{
			name:        "invalid regex pattern",
			args:        []any{"[", "text"},
			expectError: true,
		},
		{
			name:        "non-string pattern",
			args:        []any{123, "text"},
			expectError: true,
		},
		{
			name:        "non-string text",
			args:        []any{"pattern", 123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := regexMatchFunc(tt.args)

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

func TestRegexFindFunction(t *testing.T) {
	regexFindFunc, err := getRegexpFunction("regex_find")
	if err != nil {
		t.Fatalf("failed to get regex_find function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    any
		expectError bool
	}{
		{
			name:     "find match",
			args:     []any{"world", "hello world test"},
			expected: "world",
		},
		{
			name:     "no match",
			args:     []any{"xyz", "hello world"},
			expected: nil,
		},
		{
			name:     "regex pattern find",
			args:     []any{"[0-9]+", "test123world456"},
			expected: "123",
		},
		{
			name:     "find email",
			args:     []any{"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}", "Contact us at test@example.com for help"},
			expected: "test@example.com",
		},
		{
			name:        "invalid regex pattern",
			args:        []any{"[", "text"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := regexFindFunc(tt.args)

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

func TestRegexFindAllFunction(t *testing.T) {
	regexFindAllFunc, err := getRegexpFunction("regex_find_all")
	if err != nil {
		t.Fatalf("failed to get regex_find_all function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    []any
		expectError bool
	}{
		{
			name:     "find all numbers",
			args:     []any{"[0-9]+", "test123world456end789"},
			expected: []any{"123", "456", "789"},
		},
		{
			name:     "no matches",
			args:     []any{"[0-9]+", "hello world"},
			expected: []any{},
		},
		{
			name:     "find all words",
			args:     []any{"[a-zA-Z]+", "hello123world456test"},
			expected: []any{"hello", "world", "test"},
		},
		{
			name:        "invalid regex pattern",
			args:        []any{"[", "text"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := regexFindAllFunc(tt.args)

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

func TestRegexReplaceFunction(t *testing.T) {
	regexReplaceFunc, err := getRegexpFunction("regex_replace")
	if err != nil {
		t.Fatalf("failed to get regex_replace function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "simple replace",
			args:     []any{"world", "universe", "hello world"},
			expected: "hello universe",
		},
		{
			name:     "replace all numbers",
			args:     []any{"[0-9]+", "X", "test123world456"},
			expected: "testXworldX",
		},
		{
			name:     "no replacement needed",
			args:     []any{"xyz", "ABC", "hello world"},
			expected: "hello world",
		},
		{
			name:     "sanitize string",
			args:     []any{"[^a-zA-Z0-9]", "_", "hello-world!@#"},
			expected: "hello_world___",
		},
		{
			name:        "invalid regex pattern",
			args:        []any{"[", "replacement", "text"},
			expectError: true,
		},
		{
			name:        "non-string pattern",
			args:        []any{123, "replacement", "text"},
			expectError: true,
		},
		{
			name:        "non-string replacement",
			args:        []any{"pattern", 123, "text"},
			expectError: true,
		},
		{
			name:        "non-string text",
			args:        []any{"pattern", "replacement", 123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := regexReplaceFunc(tt.args)

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

func TestRegexSplitFunction(t *testing.T) {
	regexSplitFunc, err := getRegexpFunction("regex_split")
	if err != nil {
		t.Fatalf("failed to get regex_split function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    []any
		expectError bool
	}{
		{
			name:     "split by comma",
			args:     []any{",", "a,b,c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "split by spaces",
			args:     []any{"\\s+", "hello   world  test"},
			expected: []any{"hello", "world", "test"},
		},
		{
			name:     "split by numbers",
			args:     []any{"[0-9]+", "hello123world456test"},
			expected: []any{"hello", "world", "test"},
		},
		{
			name:     "no split needed",
			args:     []any{",", "hello world"},
			expected: []any{"hello world"},
		},
		{
			name:     "empty string",
			args:     []any{",", ""},
			expected: []any{""},
		},
		{
			name:        "invalid regex pattern",
			args:        []any{"[", "text"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := regexSplitFunc(tt.args)

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
