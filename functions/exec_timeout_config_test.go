package functions_test

import (
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed/functions"
)

func TestCustomExecTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping custom timeout test in short mode")
	}

	// Save original timeout
	originalTimeout := functions.DefaultExecTimeout
	defer func() {
		// Restore original timeout
		functions.DefaultExecTimeout = originalTimeout
	}()

	// Set a short timeout for testing
	functions.DefaultExecTimeout = 2 * time.Second

	execFunc, err := getExecFunction("exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	start := time.Now()

	// This should timeout in ~2 seconds due to custom DefaultExecTimeout
	result, err := execFunc([]any{"sleep", []any{"3"}})

	elapsed := time.Since(start)

	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)
	t.Logf("Elapsed: %v", elapsed)

	if err == nil {
		t.Error("expected timeout error but got nil")
	}

	// Should timeout around 2 seconds (custom timeout), not 30 seconds (default)
	if elapsed > 3*time.Second {
		t.Errorf("command took too long: %v (expected ~2s)", elapsed)
	}
	if elapsed < 1*time.Second {
		t.Errorf("command finished too quickly: %v (expected ~2s)", elapsed)
	}
}