package functions_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-cmp/cmp"
)

func TestExecFunction(t *testing.T) {
	execFunc, err := getExecFunction(t.Context(), "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    map[string]any
		expectError bool
	}{
		{
			name: "echo command",
			args: []any{"echo", []any{"hello", "world"}},
			expected: map[string]any{
				"stdout":    "hello world\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "command without args",
			args: []any{"echo", nil},
			expected: map[string]any{
				"stdout":    "\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "command with empty args array",
			args: []any{"echo", []any{}},
			expected: map[string]any{
				"stdout":    "\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "failing command",
			args: []any{"false", []any{}},
			expected: map[string]any{
				"stdout":    "",
				"stderr":    "",
				"exit_code": 1,
			},
		},
		{
			name:        "non-string command",
			args:        []any{123, []any{}},
			expectError: true,
		},
		{
			name:        "non-array args",
			args:        []any{"echo", "not-array"},
			expectError: true,
		},
		{
			name:        "non-string arg in array",
			args:        []any{"echo", []any{123}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := execFunc(tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("result is not a map: %T", result)
			}

			if diff := cmp.Diff(tt.expected, resultMap); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExecWithEnvFunction(t *testing.T) {
	execWithEnvFunc, err := getExecFunction(t.Context(), "exec_with_env")
	if err != nil {
		t.Fatalf("failed to get exec_with_env function: %v", err)
	}

	var envTestCmd, envTestArg string
	if runtime.GOOS == "windows" {
		envTestCmd = "cmd"
		envTestArg = "/C"
	} else {
		envTestCmd = "sh"
		envTestArg = "-c"
	}

	tests := []struct {
		name        string
		args        []any
		expected    map[string]any
		expectError bool
	}{
		{
			name: "command with environment variables",
			args: []any{
				envTestCmd,
				[]any{envTestArg, "echo $TEST_VAR"},
				map[string]any{"TEST_VAR": "test-value"},
			},
			expected: map[string]any{
				"stdout":    "test-value\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "command without environment variables",
			args: []any{"echo", []any{"hello"}, nil},
			expected: map[string]any{
				"stdout":    "hello\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "command with empty environment variables",
			args: []any{"echo", []any{"hello"}, map[string]any{}},
			expected: map[string]any{
				"stdout":    "hello\n",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name:        "non-string command",
			args:        []any{123, []any{}, map[string]any{}},
			expectError: true,
		},
		{
			name:        "non-array args",
			args:        []any{"echo", "not-array", map[string]any{}},
			expectError: true,
		},
		{
			name:        "non-object env_vars",
			args:        []any{"echo", []any{}, "not-object"},
			expectError: true,
		},
		{
			name: "non-string env value",
			args: []any{
				"echo",
				[]any{},
				map[string]any{"TEST_VAR": 123},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := execWithEnvFunc(tt.args)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("result is not a map: %T", result)
			}

			// Skip detailed comparison for environment variable test on Windows
			// as cmd.exe handles environment variables differently
			if runtime.GOOS == "windows" && tt.name == "command with environment variables" {
				if resultMap["exit_code"] != 0 {
					t.Errorf("expected exit_code 0, got %v", resultMap["exit_code"])
				}
				return
			}

			if diff := cmp.Diff(tt.expected, resultMap); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExecTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	// Save original timeout and restore after test
	originalTimeout := functions.DefaultExecTimeout
	defer func() {
		functions.DefaultExecTimeout = originalTimeout
	}()

	// Set shorter timeout for testing
	functions.DefaultExecTimeout = 3 * time.Second

	execFunc, err := getExecFunction(t.Context(), "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	var sleepCmd string
	var sleepArg string
	if runtime.GOOS == "windows" {
		sleepCmd = "timeout"
		sleepArg = "5"
	} else {
		sleepCmd = "sleep"
		sleepArg = "5"
	}

	t.Logf("Running command: %s %s", sleepCmd, sleepArg)

	// This should timeout (3 seconds is the test timeout, we sleep for 5 seconds)
	result, err := execFunc([]any{sleepCmd, []any{sleepArg}})

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)

	if err == nil {
		t.Error("expected timeout error but got nil")
	}
}

func TestExecTimeoutSIGKILL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("skipping SIGKILL test on Windows")
	}

	// Save original timeout and restore after test
	originalTimeout := functions.DefaultExecTimeout
	defer func() {
		functions.DefaultExecTimeout = originalTimeout
	}()

	// Set shorter timeout for testing
	functions.DefaultExecTimeout = 3 * time.Second

	execFunc, err := getExecFunction(t.Context(), "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	// This shell script ignores SIGTERM but will be killed by SIGKILL
	script := `trap '' TERM; sleep 5`

	t.Logf("Running script that ignores SIGTERM")

	// This should timeout and be killed by SIGKILL after WaitDelay
	result, err := execFunc([]any{"sh", []any{"-c", script}})

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)

	if err == nil {
		t.Error("expected timeout error but got nil")
	}
}
