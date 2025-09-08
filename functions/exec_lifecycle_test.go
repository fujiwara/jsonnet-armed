package functions_test

import (
	"context"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed/functions"
)

func TestExecContextLifecycle(t *testing.T) {
	execFunc, err := getExecFunction("exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	// Scenario 1: Set a cancelled context
	ctx1, cancel1 := context.WithCancel(context.Background())
	functions.SetExecContext(ctx1)
	cancel1() // Cancel immediately

	// Wait a bit to ensure cancellation is processed
	time.Sleep(100 * time.Millisecond)

	// Scenario 2: Reset context (simulating CLI completion)
	functions.ResetExecContext()

	// Scenario 3: Try to execute a simple command (should work, not be affected by previous cancellation)
	t.Log("Testing exec after context reset...")
	result, err := execFunc([]any{"echo", []any{"hello"}})

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)

	// This should NOT fail due to the previously cancelled context
	if err != nil {
		t.Errorf("exec failed due to previous cancelled context: %v", err)
	}

	if result != nil {
		if resultMap, ok := result.(map[string]any); ok {
			if exitCode, exists := resultMap["exit_code"]; exists && exitCode != 0 {
				t.Errorf("expected exit_code 0, got %v", exitCode)
			}
		}
	}
}

func TestMultipleCLIExecutions(t *testing.T) {
	execFunc, err := getExecFunction("exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	// Simulate first CLI execution with timeout
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel1()
	functions.SetExecContext(ctx1)

	// Wait for context to timeout
	time.Sleep(1100 * time.Millisecond)

	// Simulate second CLI execution without timeout (normal case)
	ctx2 := context.Background()
	functions.SetExecContext(ctx2)

	// This should work normally
	result, err := execFunc([]any{"echo", []any{"test"}})

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)

	if err != nil {
		t.Errorf("second CLI execution failed: %v", err)
	}
}
