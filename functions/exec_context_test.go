package functions_test

import (
	"context"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed/functions"
)

func TestExecContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping context cancellation test in short mode")
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set the context for exec functions
	functions.SetExecContext(ctx)

	execFunc, err := getExecFunction("exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	// Start a long-running command
	resultCh := make(chan execResult, 1)
	
	go func() {
		result, err := execFunc([]any{"sleep", []any{"30"}})
		resultCh <- execResult{result: result, err: err}
	}()

	// Cancel the context after 2 seconds
	time.Sleep(2 * time.Second)
	t.Logf("Cancelling context...")
	cancel()

	// Wait for the command to finish (should be cancelled)
	select {
	case res := <-resultCh:
		t.Logf("Result: %v", res.result)
		t.Logf("Error: %v", res.err)
		
		if res.err == nil {
			t.Error("expected cancellation error but got nil")
		}
		
		// Should finish quickly due to cancellation
	case <-time.After(10 * time.Second):
		t.Error("command did not finish within 10 seconds after cancellation")
	}
}

func TestExecContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping context timeout test in short mode")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	// Set the context for exec functions
	functions.SetExecContext(ctx)

	execFunc, err := getExecFunction("exec")
	if err != nil {
		t.Fatalf("failed to get exec function: %v", err)
	}

	start := time.Now()
	
	// This should be cancelled by the parent context timeout (3s) before exec timeout (30s)
	result, err := execFunc([]any{"sleep", []any{"30"}})
	
	elapsed := time.Since(start)
	
	t.Logf("Result: %v", result)
	t.Logf("Error: %v", err)
	t.Logf("Elapsed: %v", elapsed)
	
	if err == nil {
		t.Error("expected timeout error but got nil")
	}
	
	// Should finish around 3 seconds due to parent context timeout
	if elapsed > 5*time.Second {
		t.Errorf("command took too long: %v (expected around 3s)", elapsed)
	}
	if elapsed < 2*time.Second {
		t.Errorf("command finished too quickly: %v (expected around 3s)", elapsed)
	}
}

type execResult struct {
	result any
	err    error
}