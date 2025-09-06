package functions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed"
)

func TestTimeFunctions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonnet     string
		validate    func(t *testing.T, result string)
		expectError bool
	}{
		{
			name: "now returns current timestamp",
			jsonnet: `
			local now = std.native("now");
			{
				timestamp: now()
			}`,
			validate: func(t *testing.T, result string) {
				var output map[string]interface{}
				if err := json.Unmarshal([]byte(result), &output); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				timestamp, ok := output["timestamp"].(float64)
				if !ok {
					t.Fatalf("timestamp is not a number")
				}

				// Check if timestamp is close to current time (within 5 seconds)
				currentTime := float64(time.Now().Unix())
				if math.Abs(timestamp-currentTime) > 5 {
					t.Errorf("timestamp %f is not close to current time %f", timestamp, currentTime)
				}

				// Check if timestamp has fractional part (nanoseconds)
				if timestamp == math.Floor(timestamp) {
					t.Logf("Warning: timestamp has no fractional part: %f", timestamp)
				}
			},
		},
		{
			name: "time_format with RFC3339",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Use a fixed timestamp: 2024-01-15 10:30:45.123456789 UTC
				formatted: time_format(1705314645.123456789, "2006-01-02T15:04:05Z07:00")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"formatted": "2024-01-15T10:30:45Z"
				}`
				compareJSON(t, result, expected)
			},
		},
		{
			name: "time_format with custom format",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Use a fixed timestamp: 2024-01-15 10:30:45.123456789 UTC
				date: time_format(1705314645.123456789, "2006-01-02"),
				time: time_format(1705314645.123456789, "15:04:05"),
				datetime: time_format(1705314645.123456789, "2006-01-02 15:04:05"),
				with_nanos: time_format(1705314645.123456789, "2006-01-02 15:04:05.999999999")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"date": "2024-01-15",
					"time": "10:30:45",
					"datetime": "2024-01-15 10:30:45",
					"with_nanos": "2024-01-15 10:30:45.123456716"
				}`
				compareJSON(t, result, expected)
			},
		},
		{
			name: "time_format with integer timestamp",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Integer timestamp (no fractional seconds)
				formatted: time_format(1705314645, "2006-01-02T15:04:05.999999999Z")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"formatted": "2024-01-15T10:30:45Z"
				}`
				compareJSON(t, result, expected)
			},
		},
		{
			name: "time_format with invalid timestamp type",
			jsonnet: `
			local time_format = std.native("time_format");
			time_format("not a number", "2006-01-02")`,
			expectError: true,
		},
		{
			name: "time_format with invalid format type",
			jsonnet: `
			local time_format = std.native("time_format");
			time_format(1705314645, 123)`,
			expectError: true,
		},
		{
			name: "now and time_format combined",
			jsonnet: `
			local now = std.native("now");
			local time_format = std.native("time_format");
			local timestamp = now();
			{
				timestamp: timestamp,
				formatted: time_format(timestamp, "2006-01-02T15:04:05Z07:00"),
				date_only: time_format(timestamp, "2006-01-02")
			}`,
			validate: func(t *testing.T, result string) {
				var output map[string]interface{}
				if err := json.Unmarshal([]byte(result), &output); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				// Check timestamp exists and is a number
				if _, ok := output["timestamp"].(float64); !ok {
					t.Fatalf("timestamp is not a number")
				}

				// Check formatted fields exist and are strings
				if _, ok := output["formatted"].(string); !ok {
					t.Fatalf("formatted is not a string")
				}
				if _, ok := output["date_only"].(string); !ok {
					t.Fatalf("date_only is not a string")
				}
			},
		},
		{
			name: "time_format with Go format constants",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Using Go's predefined format constants names (as strings)
				kitchen: time_format(1705314645.5, "3:04PM"),
				stamp: time_format(1705314645.5, "Jan _2 15:04:05")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"kitchen": "10:30AM",
					"stamp": "Jan 15 10:30:45"
				}`
				compareJSON(t, result, expected)
			},
		},
		{
			name: "time_format with Go time constant names",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Using Go time constant names as strings
				rfc3339: time_format(1705314645.123456789, "RFC3339"),
				rfc1123: time_format(1705314645, "RFC1123"),
				dateonly: time_format(1705314645, "DateOnly"),
				timeonly: time_format(1705314645, "TimeOnly"),
				datetime: time_format(1705314645, "DateTime")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"rfc3339": "2024-01-15T10:30:45Z",
					"rfc1123": "Mon, 15 Jan 2024 10:30:45 UTC", 
					"dateonly": "2024-01-15",
					"timeonly": "10:30:45",
					"datetime": "2024-01-15 10:30:45"
				}`
				compareJSON(t, result, expected)
			},
		},
		{
			name: "time_format with mixed custom and constant formats",
			jsonnet: `
			local time_format = std.native("time_format");
			{
				// Mix of custom format and constant names
				custom: time_format(1705314645, "2006/01/02"),
				rfc3339_const: time_format(1705314645, "RFC3339"),
				rfc3339_raw: time_format(1705314645, "2006-01-02T15:04:05Z07:00")
			}`,
			validate: func(t *testing.T, result string) {
				expected := `{
					"custom": "2024/01/15",
					"rfc3339_const": "2024-01-15T10:30:45Z",
					"rfc3339_raw": "2024-01-15T10:30:45Z"
				}`
				compareJSON(t, result, expected)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with jsonnet content
			tmpDir := t.TempDir()
			jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(jsonnetFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}

			// Create CLI config with output capture
			var output bytes.Buffer
			cli := &armed.CLI{
				Filename: jsonnetFile,
			}

			// Run evaluation
			cli.SetWriter(&output)
			err := cli.Run(ctx)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Validate output
			if tt.validate != nil {
				tt.validate(t, output.String())
			}
		})
	}
}

func TestNowFunctionPrecision(t *testing.T) {
	ctx := context.Background()

	jsonnetContent := `
	local now = std.native("now");
	{
		timestamps: [now(), now(), now()]
	}`

	// Create temp file
	tmpDir := t.TempDir()
	jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
	if err := os.WriteFile(jsonnetFile, []byte(jsonnetContent), 0644); err != nil {
		t.Fatalf("failed to write jsonnet file: %v", err)
	}

	var output bytes.Buffer
	cli := &armed.CLI{
		Filename: jsonnetFile,
	}
	cli.SetWriter(&output)

	if err := cli.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string][]float64
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	timestamps := result["timestamps"]
	if len(timestamps) != 3 {
		t.Fatalf("expected 3 timestamps, got %d", len(timestamps))
	}

	// Check that timestamps are very close (called in rapid succession)
	for i := 1; i < len(timestamps); i++ {
		diff := timestamps[i] - timestamps[i-1]
		// Should be less than 0.001 second (1ms) difference
		if diff > 0.001 {
			t.Errorf("timestamps[%d] - timestamps[%d] = %f, expected < 0.001", i, i-1, diff)
		}
	}

	// Check precision - at least one should have fractional part
	hasFractional := false
	for _, ts := range timestamps {
		if ts != math.Floor(ts) {
			hasFractional = true
			break
		}
	}
	if !hasFractional {
		t.Error("no timestamp has fractional part (no nanosecond precision)")
	}
}

func TestTimeFormatWithArmedLibrary(t *testing.T) {
	ctx := context.Background()

	jsonnetContent := `
	local armed = import 'armed.libsonnet';
	{
		current: armed.time_format(armed.now(), "2006-01-02T15:04:05Z"),
		fixed: armed.time_format(1705314645.987654321, "2006-01-02 15:04:05.999999999")
	}`

	// Create temp file
	tmpDir := t.TempDir()
	jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
	if err := os.WriteFile(jsonnetFile, []byte(jsonnetContent), 0644); err != nil {
		t.Fatalf("failed to write jsonnet file: %v", err)
	}

	var output bytes.Buffer
	cli := &armed.CLI{
		Filename: jsonnetFile,
	}
	cli.SetWriter(&output)

	if err := cli.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Check current timestamp format
	if _, ok := result["current"]; !ok {
		t.Error("current field missing")
	}

	// Check fixed timestamp format
	expected := "2024-01-15 10:30:45.987654209"
	if result["fixed"] != expected {
		t.Errorf("fixed field: got %s, want %s", result["fixed"], expected)
	}
}
