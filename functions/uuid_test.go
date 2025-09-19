package functions

import (
	"regexp"
	"testing"
	"time"
)

func TestUuidV4Function(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "generate uuid v4",
			args: []any{},
		},
	}

	uuidV4Regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uuidV4Function(tt.args)
			if err != nil {
				t.Errorf("uuidV4Function() error = %v", err)
				return
			}

			uuidStr, ok := result.(string)
			if !ok {
				t.Errorf("uuidV4Function() result is not string, got %T", result)
				return
			}

			if !uuidV4Regex.MatchString(uuidStr) {
				t.Errorf("uuidV4Function() = %v, not a valid UUID v4 format", uuidStr)
			}

			// Test uniqueness - generate multiple UUIDs and ensure they're different
			result2, err := uuidV4Function(tt.args)
			if err != nil {
				t.Errorf("uuidV4Function() second call error = %v", err)
				return
			}

			uuidStr2, ok := result2.(string)
			if !ok {
				t.Errorf("uuidV4Function() second result is not string, got %T", result2)
				return
			}

			if uuidStr == uuidStr2 {
				t.Errorf("uuidV4Function() generated identical UUIDs: %v", uuidStr)
			}
		})
	}
}

func TestUuidV7Function(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "generate uuid v7",
			args: []any{},
		},
	}

	uuidV7Regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uuidV7Function(tt.args)
			if err != nil {
				t.Errorf("uuidV7Function() error = %v", err)
				return
			}

			uuidStr, ok := result.(string)
			if !ok {
				t.Errorf("uuidV7Function() result is not string, got %T", result)
				return
			}

			if !uuidV7Regex.MatchString(uuidStr) {
				t.Errorf("uuidV7Function() = %v, not a valid UUID v7 format", uuidStr)
			}

			// Test time ordering - generate UUIDs with small delay and ensure they're ordered
			time.Sleep(1 * time.Millisecond)
			result2, err := uuidV7Function(tt.args)
			if err != nil {
				t.Errorf("uuidV7Function() second call error = %v", err)
				return
			}

			uuidStr2, ok := result2.(string)
			if !ok {
				t.Errorf("uuidV7Function() second result is not string, got %T", result2)
				return
			}

			if uuidStr == uuidStr2 {
				t.Errorf("uuidV7Function() generated identical UUIDs: %v", uuidStr)
			}

			// UUID v7 should be lexicographically sortable by time
			if uuidStr >= uuidStr2 {
				t.Errorf("uuidV7Function() UUIDs not properly time-ordered: %v >= %v", uuidStr, uuidStr2)
			}
		})
	}
}
