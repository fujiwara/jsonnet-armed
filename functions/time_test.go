package functions_test

import (
	"math"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-cmp/cmp"
)

func TestNowFunction(t *testing.T) {
	nowFunc := functions.TimeFunctions[0].Func // now function

	// Call the function
	result, err := nowFunc([]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result is a float64
	timestamp, ok := result.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", result)
	}

	// Verify timestamp is reasonable (within 1 second of current time)
	now := float64(time.Now().UnixNano()) / float64(time.Second)
	if math.Abs(timestamp-now) > 1.0 {
		t.Errorf("timestamp seems too far from current time: got %f, expected around %f", timestamp, now)
	}
}

func TestTimeFormatFunction(t *testing.T) {
	timeFormatFunc := functions.TimeFunctions[1].Func // time_format function

	// Fixed timestamp for consistent testing
	// 2024-01-15 10:30:45 UTC
	timestamp := float64(1705314645)

	tests := []struct {
		name        string
		args        []any
		expected    string
		expectError bool
	}{
		{
			name:     "RFC3339 format",
			args:     []any{timestamp, "RFC3339"},
			expected: "2024-01-15T10:30:45Z",
		},
		{
			name:     "DateTime format",
			args:     []any{timestamp, "DateTime"},
			expected: "2024-01-15 10:30:45",
		},
		{
			name:     "DateOnly format",
			args:     []any{timestamp, "DateOnly"},
			expected: "2024-01-15",
		},
		{
			name:     "TimeOnly format",
			args:     []any{timestamp, "TimeOnly"},
			expected: "10:30:45",
		},
		{
			name:     "custom format",
			args:     []any{timestamp, "2006/01/02 15:04:05"},
			expected: "2024/01/15 10:30:45",
		},
		{
			name:     "year-month format",
			args:     []any{timestamp, "2006-01"},
			expected: "2024-01",
		},
		{
			name:        "non-number timestamp",
			args:        []any{"not-a-number", "RFC3339"},
			expectError: true,
		},
		{
			name:        "non-string format",
			args:        []any{timestamp, 123},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := timeFormatFunc(tt.args)

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

func TestTimeFormatWithNanoseconds(t *testing.T) {
	timeFormatFunc := functions.TimeFunctions[1].Func // time_format function

	// Timestamp with fractional seconds (nanosecond precision)
	timestamp := 1705314645.123456789

	result, err := timeFormatFunc([]any{timestamp, "RFC3339Nano"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The result should include nanoseconds
	expected := "2024-01-15T10:30:45.123456716Z"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("result mismatch (-want +got):\n%s", diff)
	}
}
