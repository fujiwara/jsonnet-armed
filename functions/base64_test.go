package functions_test

import (
	"testing"

	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-cmp/cmp"
)

func TestBase64Function(t *testing.T) {
	base64Func := functions.Base64Functions[0].Func // base64 function

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "simple string",
			args:     []any{"Hello, World!"},
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "empty string",
			args:     []any{""},
			expected: "",
		},
		{
			name:     "special characters",
			args:     []any{"!@#$%^&*()"},
			expected: "IUAjJCVeJiooKQ==",
		},
		{
			name:     "unicode string",
			args:     []any{"こんにちは"},
			expected: "44GT44KT44Gr44Gh44Gv",
		},
		{
			name:        "non-string input",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64Func(tt.args)

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

func TestBase64URLFunction(t *testing.T) {
	base64urlFunc := functions.Base64Functions[1].Func // base64url function

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "simple string",
			args:     []any{"Hello, World!"},
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "URL-unsafe characters",
			args:     []any{"??>>"},
			expected: "Pz8-Pg==",
		},
		{
			name:        "non-string input",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64urlFunc(tt.args)

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
