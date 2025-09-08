package functions_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEnvFunction(t *testing.T) {
	envFunc, err := getEnvFunction("env")
	if err != nil {
		t.Fatalf("failed to get env function: %v", err)
	}

	// Set test environment variable
	t.Setenv("TEST_ENV_VAR", "test-value")

	tests := []struct {
		name        string
		args        []any
		expected    any
		expectError bool
	}{
		{
			name:     "existing variable",
			args:     []any{"TEST_ENV_VAR", "default"},
			expected: "test-value",
		},
		{
			name:     "non-existing variable returns default",
			args:     []any{"TEST_UNSET_VAR", "default-value"},
			expected: "default-value",
		},
		{
			name:     "empty string as default",
			args:     []any{"TEST_UNSET_VAR", ""},
			expected: "",
		},
		{
			name:     "null as default",
			args:     []any{"TEST_UNSET_VAR", nil},
			expected: nil,
		},
		{
			name:     "object as default",
			args:     []any{"TEST_UNSET_VAR", map[string]any{"key": "value"}},
			expected: map[string]any{"key": "value"},
		},
		{
			name:        "non-string name",
			args:        []any{123, "default"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := envFunc(tt.args)

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

func TestEnvFunctionWithEmptyEnvVar(t *testing.T) {
	envFunc, err := getEnvFunction("env")
	if err != nil {
		t.Fatalf("failed to get env function: %v", err)
	}

	// Set environment variable to empty string
	t.Setenv("TEST_EMPTY_VAR", "")

	result, err := envFunc([]any{"TEST_EMPTY_VAR", "default"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty environment variable should return default
	if result != "default" {
		t.Errorf("expected 'default', got %v", result)
	}
}

func TestMustEnvFunction(t *testing.T) {
	mustEnvFunc, err := getEnvFunction("must_env")
	if err != nil {
		t.Fatalf("failed to get must_env function: %v", err)
	}

	// Set test environment variable
	t.Setenv("TEST_ENV_VAR", "test-value")

	tests := []struct {
		name        string
		args        []any
		expected    any
		expectError bool
	}{
		{
			name:     "existing variable",
			args:     []any{"TEST_ENV_VAR"},
			expected: "test-value",
		},
		{
			name:        "non-existing variable",
			args:        []any{"TEST_UNSET_VAR"},
			expectError: true,
		},
		{
			name:        "non-string name",
			args:        []any{123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mustEnvFunc(tt.args)

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

func TestMustEnvWithEmptyEnvVar(t *testing.T) {
	mustEnvFunc, err := getEnvFunction("must_env")
	if err != nil {
		t.Fatalf("failed to get must_env function: %v", err)
	}

	// Set environment variable to empty string
	t.Setenv("TEST_EMPTY_VAR", "")

	result, err := mustEnvFunc([]any{"TEST_EMPTY_VAR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty environment variable should still be returned (it exists)
	if result != "" {
		t.Errorf("expected empty string, got %v", result)
	}
}

func TestEnvParseFunction(t *testing.T) {
	envParseFunc, err := getEnvFunction("env_parse")
	if err != nil {
		t.Fatalf("failed to get env_parse function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    any
		expectError bool
	}{
		{
			name: "simple env file",
			args: []any{"KEY1=value1\nKEY2=value2\n"},
			expected: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "env file with comments",
			args: []any{"# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2\n"},
			expected: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "env file with empty lines",
			args: []any{"KEY1=value1\n\nKEY2=value2\n\n"},
			expected: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "env file with quotes",
			args: []any{"KEY1=\"value with spaces\"\nKEY2='single quotes'\n"},
			expected: map[string]any{
				"KEY1": "value with spaces",
				"KEY2": "single quotes",
			},
		},
		{
			name: "env file with export keyword",
			args: []any{"export KEY1=value1\nexport KEY2=value2\n"},
			expected: map[string]any{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:     "empty content",
			args:     []any{""},
			expected: map[string]any{},
		},
		{
			name:        "non-string content",
			args:        []any{123},
			expectError: true,
		},
		{
			name:        "invalid env format",
			args:        []any{"INVALID LINE WITHOUT EQUALS\n"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := envParseFunc(tt.args)

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
