package functions_test

import (
	"context"
	"testing"
	"time"
)

func TestExecContextLifecycle(t *testing.T) {

	// Scenario 1: Set a cancelled context
	ctx1, cancel1 := context.WithCancel(context.Background())
	execFunc, err := getExecFunction(ctx1, "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}
	cancel1() // Cancel immediately

	// Wait a bit to ensure cancellation is processed
	time.Sleep(100 * time.Millisecond)

	ctx2 := t.Context()
	execFunc, err = getExecFunction(ctx2, "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

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
	// Simulate first CLI execution with timeout
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	execFunc, err := getExecFunction(ctx1, "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}
	defer cancel1()

	// Wait for context to timeout
	time.Sleep(1100 * time.Millisecond)

	// Simulate second CLI execution without timeout (normal case)
	ctx2 := t.Context()
	execFunc, err = getExecFunction(ctx2, "exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	// This should work normally
	result, err := execFunc([]any{"echo", []any{"test"}})

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)

	if err != nil {
		t.Errorf("second CLI execution failed: %v", err)
	}
}
