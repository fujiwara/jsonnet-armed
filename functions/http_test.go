package functions

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHttpFunctions(t *testing.T) {
	// Setup context with version
	ctx := context.WithValue(context.Background(), versionKey, "v0.0.7-test")
	httpFuncs := GenerateHttpFunctions(ctx)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Test-Header", "test-value")
			// Echo User-Agent for testing
			if ua := r.Header.Get("User-Agent"); ua != "" {
				w.Header().Set("X-User-Agent", ua)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"message": "test"}`)
		case "/multiple-headers":
			w.Header().Add("Set-Cookie", "cookie1=value1")
			w.Header().Add("Set-Cookie", "cookie2=value2")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "multiple headers")
		case "/post":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"status": "created"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		function    string
		args        []any
		expected    map[string]any
		expectError bool
	}{
		{
			name:     "http_get with valid URL",
			function: "http_get",
			args:     []any{server.URL + "/test", nil},
			expected: map[string]any{
				"status_code": 200,
				"status":      "200 OK",
				"headers": map[string]any{
					"Content-Type":    "application/json",
					"X-Test-Header":   "test-value",
					"X-User-Agent":    "jsonnet-armed/v0.0.7-test",
					"Content-Length":  "19",
				},
				"body": `{"message": "test"}`,
			},
		},
		{
			name:     "http_get with custom headers",
			function: "http_get",
			args: []any{
				server.URL + "/test",
				map[string]any{"User-Agent": "jsonnet-armed"},
			},
			expected: map[string]any{
				"status_code": 200,
				"status":      "200 OK",
				"headers": map[string]any{
					"Content-Type":    "application/json",
					"X-Test-Header":   "test-value",
					"X-User-Agent":    "jsonnet-armed", // Custom User-Agent
					"Content-Length":  "19",
				},
				"body": `{"message": "test"}`,
			},
		},
		{
			name:     "http_get with multiple header values",
			function: "http_get",
			args:     []any{server.URL + "/multiple-headers", nil},
			expected: map[string]any{
				"status_code": 200,
				"status":      "200 OK",
				"headers": map[string]any{
					"Set-Cookie":     []string{"cookie1=value1", "cookie2=value2"},
					"Content-Length": "16",
				},
				"body": "multiple headers",
			},
		},
		{
			name:     "http_request POST with body",
			function: "http_request",
			args: []any{
				"POST",
				server.URL + "/post",
				map[string]any{"Content-Type": "application/json"},
				`{"data": "test"}`,
			},
			expected: map[string]any{
				"status_code": 201,
				"status":      "201 Created",
				"headers": map[string]any{
					"Content-Type":   "application/json",
					"Content-Length": "21",
				},
				"body": `{"status": "created"}`,
			},
		},
		{
			name:        "http_get with invalid URL",
			function:    "http_get",
			args:        []any{"invalid-url", nil},
			expectError: true,
		},
		{
			name:        "http_get with non-string URL",
			function:    "http_get",
			args:        []any{123, nil},
			expectError: true,
		},
		{
			name:        "http_request with non-string method",
			function:    "http_request",
			args:        []any{123, server.URL + "/test", nil, nil},
			expectError: true,
		},
		{
			name:        "http_request with invalid headers",
			function:    "http_request",
			args:        []any{"GET", server.URL + "/test", "invalid-headers", nil},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fn func([]any) (any, error)
			switch tt.function {
			case "http_get":
				fn = httpFuncs["http_get"].Func
			case "http_request":
				fn = httpFuncs["http_request"].Func
			default:
				t.Fatalf("Unknown function: %s", tt.function)
			}

			result, err := fn(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Errorf("Expected result to be map[string]any, got %T", result)
				return
			}

			// Compare status_code and status
			if diff := cmp.Diff(tt.expected["status_code"], resultMap["status_code"]); diff != "" {
				t.Errorf("status_code mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected["status"], resultMap["status"]); diff != "" {
				t.Errorf("status mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected["body"], resultMap["body"]); diff != "" {
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}

			// Compare headers (excluding dynamic headers like Date)
			expectedHeaders := tt.expected["headers"].(map[string]any)
			resultHeaders := resultMap["headers"].(map[string]any)

			for key, expectedValue := range expectedHeaders {
				if resultValue, exists := resultHeaders[key]; exists {
					if diff := cmp.Diff(expectedValue, resultValue); diff != "" {
						t.Errorf("header %s mismatch (-want +got):\n%s", key, diff)
					}
				} else {
					t.Errorf("Expected header %s not found in result", key)
				}
			}
		})
	}
}
