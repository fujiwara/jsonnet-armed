package armed_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/google/go-cmp/cmp"
)

func TestIntegrationExamples(t *testing.T) {
	ctx := context.Background()

	// Set test environment variables
	t.Setenv("HOME", "/tmp")
	t.Setenv("USER", "testuser")

	tests := []struct {
		name     string
		jsonnet  string
		expected map[string]interface{}
		extStr   map[string]string
		extCode  map[string]string
		setup    func(t *testing.T) string // returns tmpDir if needed
		cleanup  func(tmpDir string)
	}{
		{
			name: "README env example",
			jsonnet: `
			local env = std.native("env");
			{
				// Returns the value of HOME environment variable, or "/tmp" if not set
				home: env("HOME", "/tmp"),
				// Can use any JSON value as default
				config: env("CONFIG", { debug: false })
			}`,
			expected: map[string]interface{}{
				"home":   "/tmp",
				"config": map[string]interface{}{"debug": false},
			},
		},
		{
			name: "README base64 example",
			jsonnet: `
			local base64 = std.native("base64");
			local base64url = std.native("base64url");
			{
				// Standard Base64 encoding
				encoded: base64("Hello, World!"),
				empty: base64(""),
				// URL-safe Base64 encoding  
				url_safe: base64url("??>>"),
				// Encoding with special characters
				unicode: base64("こんにちは世界")
			}`,
			expected: map[string]interface{}{
				"encoded":  "SGVsbG8sIFdvcmxkIQ==",
				"empty":    "",
				"url_safe": "Pz8-Pg==",
				"unicode":  "44GT44KT44Gr44Gh44Gv5LiW55WM",
			},
		},
		{
			name: "README hash example",
			jsonnet: `
			local md5 = std.native("md5");
			local sha1 = std.native("sha1");
			local sha256 = std.native("sha256");
			{
				// String hash functions
				md5_hash: md5("hello"),
				sha1_hash: sha1("hello"),
				sha256_hash: sha256("hello"),
				// Can be used with variables
				short_hash: std.substr(sha256("data"), 0, 8)
			}`,
			expected: map[string]interface{}{
				"md5_hash":    "5d41402abc4b2a76b9719d911017c592",
				"sha1_hash":   "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
				"sha256_hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				"short_hash":  "3a6eb079",
			},
		},
		{
			name: "README file functions example",
			jsonnet: `
			local file_content = std.native("file_content");
			local file_stat = std.native("file_stat");
			local file_exists = std.native("file_exists");
			{
				// Check file existence
				readme_exists: file_exists("README.md"),
				fake_exists: file_exists("/fake/path/file.txt"),
				// Conditional file reading
				config: if file_exists("test_config.json")
						then std.parseJson(file_content("test_config.json"))
						else {"default": "config"},
				// File info
				file_info: if file_exists("test_file.txt") then {
					name: file_stat("test_file.txt").name,
					size: file_stat("test_file.txt").size,
					is_dir: file_stat("test_file.txt").is_dir
				} else null
			}`,
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create test files
				configPath := filepath.Join(tmpDir, "test_config.json")
				if err := os.WriteFile(configPath, []byte(`{"app": "test", "version": "1.0"}`), 0644); err != nil {
					t.Fatalf("failed to create test config: %v", err)
				}

				testFilePath := filepath.Join(tmpDir, "test_file.txt")
				if err := os.WriteFile(testFilePath, []byte("hello world"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				// Change working directory to tmpDir so relative paths work
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				t.Cleanup(func() { os.Chdir(oldWd) })

				return tmpDir
			},
			expected: map[string]interface{}{
				"readme_exists": false, // README.md won't exist in temp dir
				"fake_exists":   false,
				"config": map[string]interface{}{
					"app":     "test",
					"version": "1.0",
				},
				"file_info": map[string]interface{}{
					"name":   "test_file.txt",
					"size":   float64(11), // "hello world" is 11 bytes
					"is_dir": false,
				},
			},
		},
		{
			name: "README armed.libsonnet example",
			jsonnet: `
			local armed = import 'armed.libsonnet';
			{
				sha256_test: armed.sha256('test'),
				env_test: armed.env('USER', 'default_user'),
			}`,
			expected: map[string]interface{}{
				"sha256_test": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				"env_test":    "testuser",
			},
		},
		{
			name: "README external variables example",
			jsonnet: `
			local env = std.native("env");
			local md5 = std.native("md5");
			local sha256 = std.native("sha256");
			{
				// External variables
				environment: std.extVar("env"),
				region: std.extVar("region"),
				replicas: std.extVar("replicas"),
				debug: std.extVar("debug"),
				// Environment variables
				home_dir: env("HOME", "/home/user"),
				// Hash functions
				config_hash: sha256(std.extVar("env") + std.extVar("region")),
				short_id: md5(std.extVar("instance_id"))[0:8]
			}`,
			extStr: map[string]string{
				"env":         "production",
				"region":      "us-west-2",
				"instance_id": "i-1234567890abcdef0",
			},
			extCode: map[string]string{
				"replicas": "3",
				"debug":    "false",
			},
			expected: map[string]interface{}{
				"environment": "production",
				"region":      "us-west-2",
				"replicas":    float64(3),
				"debug":       false,
				"home_dir":    "/tmp",
				"config_hash": "dd7a6b797440e1f28ca030baf32be3b604ef1669bcca017eed2c4e65f43ace7a",
				"short_id":    "c65b83a2",
			},
		},
		{
			name: "README exec functions example",
			jsonnet: `
			local exec = std.native("exec");
			local exec_with_env = std.native("exec_with_env");
			{
				// Basic command execution
				hello: exec("echo", ["Hello, World!"]),
				
				// Check exit code for success
				success: exec("true", []).exit_code == 0,
				failure: exec("false", []).exit_code == 0,
				
				// Command with environment variables
				custom_env: exec_with_env("sh", ["-c", "echo $TEST_VAR"], {
					"TEST_VAR": "test-value"
				}),
				
				// Safe command execution
				date_result: {
					local result = exec("date", ["+%Y"]),
					success: result.exit_code == 0,
					year_length: if result.exit_code == 0 then std.length(std.strReplace(result.stdout, "\n", "")) else 0
				}
			}`,
			expected: map[string]interface{}{
				"hello": map[string]interface{}{
					"stdout":    "Hello, World!\n",
					"stderr":    "",
					"exit_code": float64(0),
				},
				"success": true,
				"failure": false,
				"custom_env": map[string]interface{}{
					"stdout":    "test-value\n",
					"stderr":    "",
					"exit_code": float64(0),
				},
				"date_result": map[string]interface{}{
					"success":     true,
					"year_length": float64(4), // YYYY format should be 4 characters
				},
			},
		},
		{
			name: "README armed.libsonnet with exec example",
			jsonnet: `
			local armed = import 'armed.libsonnet';
			{
				sha256_test: armed.sha256('test'),
				env_test: armed.env('USER', 'default_user'),
				command_result: armed.exec('echo', ['Hello, World!']),
				date_check: {
					local result = armed.exec('date', ['+%Y-%m-%d']),
					success: result.exit_code == 0,
					has_output: std.length(result.stdout) > 0
				}
			}`,
			expected: map[string]interface{}{
				"sha256_test": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
				"env_test":    "testuser",
				"command_result": map[string]interface{}{
					"stdout":    "Hello, World!\n",
					"stderr":    "",
					"exit_code": float64(0),
				},
				"date_check": map[string]interface{}{
					"success":    true,
					"has_output": true,
				},
			},
		},
		{
			name: "env_parse function",
			jsonnet: `
			local env_parse = std.native("env_parse");
			local file_content = std.native("file_content");
			{
				// Parse .env file content
				env_from_file: env_parse(file_content("testdata/test.env")),
				
				// Parse inline env format string
				inline_env: env_parse("KEY1=value1\nKEY2=value2\n# comment\nKEY3=value3"),
				
				// Use parsed env values
				database_url: env_parse(file_content("testdata/test.env")).DATABASE_URL,
				api_key: env_parse(file_content("testdata/test.env")).API_KEY,
				
				// Handle empty content
				empty: env_parse(""),
				
				// Parse with various formats
				mixed_format: env_parse("SIMPLE=value\nexport EXPORTED=exported_value\n# Comment line\n\nQUOTED=\"quoted value\"\n")
			}`,
			expected: map[string]interface{}{
				"env_from_file": map[string]interface{}{
					"DATABASE_URL":  "postgres://user:pass@localhost/db",
					"API_KEY":       "secret-key-123",
					"DEBUG":         "true",
					"PORT":          "8080",
					"TIMEOUT":       "30s",
					"MESSAGE":       "Hello, World!",
					"SINGLE_QUOTES": "Another message",
					"LAST_VAR":      "final_value",
				},
				"inline_env": map[string]interface{}{
					"KEY1": "value1",
					"KEY2": "value2",
					"KEY3": "value3",
				},
				"database_url": "postgres://user:pass@localhost/db",
				"api_key":      "secret-key-123",
				"empty":        map[string]interface{}{},
				"mixed_format": map[string]interface{}{
					"SIMPLE":   "value",
					"EXPORTED": "exported_value",
					"QUOTED":   "quoted value",
				},
			},
		},
		{
			name: "HTTP functions example",
			jsonnet: `
			local http_get = std.native("http_get");
			local http_request = std.native("http_request");
			local server_url = std.extVar("SERVER_URL");
			{
				// Simple GET request
				get_response: http_get(server_url + "/get", null),
				// POST request with headers and body
				post_response: http_request("POST", server_url + "/post",
					{"Content-Type": "application/json"},
					'{"test": "data"}')
			}`,
			expected: map[string]interface{}{
				"get_response": map[string]interface{}{
					"status_code": float64(200),
					"status":      "200 OK",
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
					"body": `{"message": "get"}`,
				},
				"post_response": map[string]interface{}{
					"status_code": float64(201),
					"status":      "201 Created",
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
					"body": `{"message": "post"}`,
				},
			},
			setup: func(t *testing.T) string {
				// Start test HTTP server
				return startTestHTTPServer(t)
			},
			cleanup: func(tmpDir string) {
				// Cleanup is handled by test server context
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpDir string
			var serverURL string
			if tt.setup != nil {
				result := tt.setup(t)
				if strings.HasPrefix(result, "http") {
					// setup returned server URL
					serverURL = result
					tmpDir = t.TempDir()
				} else {
					// setup returned tmpDir
					tmpDir = result
				}
			}
			if tt.cleanup != nil {
				defer tt.cleanup(tmpDir)
			}

			// Create temp file with jsonnet content
			if tmpDir == "" {
				tmpDir = t.TempDir()
			}
			jsonnetFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(jsonnetFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("failed to write jsonnet file: %v", err)
			}

			// Create CLI config with output capture
			var output bytes.Buffer
			extStr := tt.extStr
			if extStr == nil {
				extStr = make(map[string]string)
			}
			if serverURL != "" {
				extStr["SERVER_URL"] = serverURL
			}

			cli := &armed.CLI{
				Filename: jsonnetFile,
				ExtStr:   extStr,
				ExtCode:  tt.extCode,
			}
			cli.SetWriter(&output)

			// Run evaluation
			err := cli.Run(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse output JSON
			var result map[string]interface{}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatalf("failed to parse result JSON: %v\nOutput: %s", err, output.String())
			}

			// Compare results - handle timestamps specially
			if diff := compareWithTimeStampTolerance(tt.expected, result); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// compareWithTimeStampTolerance compares two maps but ignores timestamp fields
// since they are dynamic and can't be predicted exactly
func compareWithTimeStampTolerance(expected, actual map[string]interface{}) string {
	// Make copies to avoid modifying originals
	expectedCopy := deepCopyMap(expected)
	actualCopy := deepCopyMap(actual)

	// Remove or normalize timestamp fields
	normalizeTimestamps(expectedCopy)
	normalizeTimestamps(actualCopy)

	return cmp.Diff(expectedCopy, actualCopy)
}

func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			copy[k] = deepCopyMap(val)
		default:
			copy[k] = val
		}
	}
	return copy
}

func normalizeTimestamps(m map[string]interface{}) {
	timestampPattern := regexp.MustCompile(`^(timestamp|time|mod_time|now)`)

	for k, v := range m {
		if timestampPattern.MatchString(k) {
			// For timestamp fields, just verify they're numeric/string
			switch v.(type) {
			case float64, string:
				// Keep as is - these are likely valid timestamps
			}
		} else if k == "headers" && v != nil {
			// Remove dynamic HTTP headers
			if headersMap, ok := v.(map[string]interface{}); ok {
				delete(headersMap, "Date")
				delete(headersMap, "Content-Length") // Can vary based on server implementation
			}
		} else if subMap, ok := v.(map[string]interface{}); ok {
			normalizeTimestamps(subMap)
		}
	}
}

func TestIntegrationCache(t *testing.T) {
	ctx := context.Background()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "cache_test.jsonnet")
	jsonnet := `{ timestamp: std.extVar("ts"), value: "cached" }`
	if err := os.WriteFile(testFile, []byte(jsonnet), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// First evaluation with cache
	cli1 := &armed.CLI{
		Filename: testFile,
		ExtStr:   map[string]string{"ts": "first"},
		Cache:    time.Hour, // Long cache duration
	}
	var buf1 bytes.Buffer
	cli1.SetWriter(&buf1)

	if err := cli1.Run(ctx); err != nil {
		t.Fatalf("First evaluation failed: %v", err)
	}

	firstResult := buf1.String()
	var firstJSON map[string]interface{}
	if err := json.Unmarshal([]byte(firstResult), &firstJSON); err != nil {
		t.Fatalf("Failed to parse first result: %v", err)
	}

	// Verify first result
	if firstJSON["timestamp"] != "first" {
		t.Errorf("Expected timestamp 'first', got %v", firstJSON["timestamp"])
	}

	// Second evaluation with different extStr but same cache
	cli2 := &armed.CLI{
		Filename: testFile,
		ExtStr:   map[string]string{"ts": "second"}, // Different value
		Cache:    time.Hour,                         // Same cache duration
	}
	var buf2 bytes.Buffer
	cli2.SetWriter(&buf2)

	if err := cli2.Run(ctx); err != nil {
		t.Fatalf("Second evaluation failed: %v", err)
	}

	secondResult := buf2.String()
	var secondJSON map[string]interface{}
	if err := json.Unmarshal([]byte(secondResult), &secondJSON); err != nil {
		t.Fatalf("Failed to parse second result: %v", err)
	}

	// Different extStr should generate different cache key
	if secondJSON["timestamp"] != "second" {
		t.Errorf("Expected timestamp 'second', got %v", secondJSON["timestamp"])
	}

	// Third evaluation with same parameters as first (cache hit)
	cli3 := &armed.CLI{
		Filename: testFile,
		ExtStr:   map[string]string{"ts": "first"}, // Same as first
		Cache:    time.Hour,
	}
	var buf3 bytes.Buffer
	cli3.SetWriter(&buf3)

	if err := cli3.Run(ctx); err != nil {
		t.Fatalf("Third evaluation failed: %v", err)
	}

	thirdResult := buf3.String()

	// Should get exact same result as first (from cache)
	if diff := cmp.Diff(firstResult, thirdResult); diff != "" {
		t.Errorf("Cache hit should return identical result (-first +third):\n%s", diff)
	}
}

func TestIntegrationStaleCache(t *testing.T) {
	ctx := context.Background()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "stale_test.jsonnet")
	// Use must_env which will fail if the env var doesn't exist
	validJsonnet := `
		local must_env = std.native("must_env");
		{
			value: "success",
			timestamp: std.extVar("ts"),
			env_check: must_env("TEST_STALE_CACHE_VAR")
		}`
	if err := os.WriteFile(testFile, []byte(validJsonnet), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Set the environment variable for the first evaluation
	t.Setenv("TEST_STALE_CACHE_VAR", "env_value")

	// First evaluation to create cache
	cli1 := &armed.CLI{
		Filename: testFile,
		ExtStr:   map[string]string{"ts": "cached"},
		Cache:    50 * time.Millisecond,  // Short cache TTL
		Stale:    500 * time.Millisecond, // Longer stale TTL
	}
	var buf1 bytes.Buffer
	cli1.SetWriter(&buf1)

	if err := cli1.Run(ctx); err != nil {
		t.Fatalf("First evaluation failed: %v", err)
	}

	firstResult := buf1.String()
	expected := map[string]interface{}{
		"value":     "success",
		"timestamp": "cached",
		"env_check": "env_value",
	}

	var firstJSON map[string]interface{}
	if err := json.Unmarshal([]byte(firstResult), &firstJSON); err != nil {
		t.Fatalf("Failed to parse first result: %v", err)
	}

	if diff := cmp.Diff(expected, firstJSON); diff != "" {
		t.Errorf("First result mismatch (-want +got):\n%s", diff)
	}

	// Wait for cache to expire but stay within stale period
	time.Sleep(100 * time.Millisecond)

	// Unset the environment variable to cause must_env to fail
	os.Unsetenv("TEST_STALE_CACHE_VAR")

	// Capture slog output to verify warning message
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	slog.SetDefault(logger)

	// Second evaluation should use stale cache due to evaluation error
	cli2 := &armed.CLI{
		Filename: testFile,
		ExtStr:   map[string]string{"ts": "cached"},
		Cache:    50 * time.Millisecond,
		Stale:    500 * time.Millisecond,
	}
	var buf2 bytes.Buffer
	cli2.SetWriter(&buf2)

	if err := cli2.Run(ctx); err != nil {
		t.Fatalf("Second evaluation should succeed using stale cache but failed: %v", err)
	}

	secondResult := buf2.String()

	// Should get same result as first evaluation (from stale cache)
	if diff := cmp.Diff(firstResult, secondResult); diff != "" {
		t.Errorf("Stale cache should return same result (-first +second):\n%s", diff)
	}

	// Verify warning message was logged
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Evaluation failed, using stale cache") {
		t.Errorf("Expected warning message about stale cache usage, got: %s", logOutput)
	}
}

// startTestHTTPServer starts a test HTTP server for integration tests
func startTestHTTPServer(t *testing.T) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"message": "get"}`)
		case "/post":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"message": "post"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Use server.URL instead of fixed port
	t.Cleanup(server.Close)

	return server.URL
}
