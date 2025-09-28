package functions_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestJQFunction(t *testing.T) {
	jqFunc, err := getJQFunction("jq")
	if err != nil {
		t.Fatalf("failed to get jq function: %v", err)
	}

	tests := []struct {
		name        string
		args        []any
		expected    any
		expectError bool
	}{
		{
			name:     "simple filter",
			args:     []any{".Foo", map[string]any{"Foo": "Bar"}},
			expected: "Bar",
		},
		{
			name:     "empty result",
			args:     []any{".Foo", map[string]any{"Bar": nil}},
			expected: nil,
		},
		{
			name:     "multiple results",
			args:     []any{".[]", []any{float64(1), float64(2), float64(3)}},
			expected: []any{float64(1), float64(2), float64(3)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jqFunc(tt.args)

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
