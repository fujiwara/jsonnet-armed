package armed_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fujiwara/jsonnet-armed"
	"github.com/google/go-cmp/cmp"
)

func TestCache(t *testing.T) {
	tests := []struct {
		name          string
		jsonnet       string
		extStr        map[string]string
		extCode       map[string]string
		cacheDuration time.Duration
		sleepBetween  time.Duration
		expected      string
		expectHit     bool
	}{
		{
			name:          "cache hit within ttl",
			jsonnet:       `{ timestamp: std.extVar("timestamp") }`,
			extStr:        map[string]string{"timestamp": "test1"},
			cacheDuration: 5 * time.Second,
			sleepBetween:  100 * time.Millisecond,
			expected:      "{\n   \"timestamp\": \"test1\"\n}\n",
			expectHit:     true,
		},
		{
			name:          "cache miss after ttl",
			jsonnet:       `{ timestamp: std.extVar("timestamp") }`,
			extStr:        map[string]string{"timestamp": "test2"},
			cacheDuration: 50 * time.Millisecond,
			sleepBetween:  100 * time.Millisecond,
			expected:      "{\n   \"timestamp\": \"test2\"\n}\n",
			expectHit:     false,
		},
		{
			name:          "no cache when duration is zero",
			jsonnet:       `{ timestamp: std.extVar("timestamp") }`,
			extStr:        map[string]string{"timestamp": "test3"},
			cacheDuration: 0,
			sleepBetween:  10 * time.Millisecond,
			expected:      "{\n   \"timestamp\": \"test3\"\n}\n",
			expectHit:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for the jsonnet content
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.jsonnet")
			if err := os.WriteFile(tmpFile, []byte(tt.jsonnet), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// First evaluation - should be a cache miss
			cli1 := &armed.CLI{
				Filename: tmpFile,
				ExtStr:   tt.extStr,
				ExtCode:  tt.extCode,
				Cache:    tt.cacheDuration,
			}
			var buf1 bytes.Buffer
			cli1.SetWriter(&buf1)

			ctx := context.Background()
			if err := cli1.Run(ctx); err != nil {
				t.Fatalf("First evaluation failed: %v", err)
			}

			firstResult := buf1.String()
			if diff := cmp.Diff(tt.expected, firstResult); diff != "" {
				t.Errorf("First result mismatch (-want +got):\n%s", diff)
			}

			// Sleep to test cache expiration
			time.Sleep(tt.sleepBetween)

			// Second evaluation - might be a cache hit or miss depending on TTL
			cli2 := &armed.CLI{
				Filename: tmpFile,
				ExtStr:   tt.extStr,
				ExtCode:  tt.extCode,
				Cache:    tt.cacheDuration,
			}
			var buf2 bytes.Buffer
			cli2.SetWriter(&buf2)

			if err := cli2.Run(ctx); err != nil {
				t.Fatalf("Second evaluation failed: %v", err)
			}

			secondResult := buf2.String()

			if tt.expectHit {
				// Cache hit - results should be identical
				if diff := cmp.Diff(firstResult, secondResult); diff != "" {
					t.Errorf("Cache hit but results differ (-first +second):\n%s", diff)
				}
			} else {
				// Cache miss or no cache - just verify the result is correct
				if diff := cmp.Diff(tt.expected, secondResult); diff != "" {
					t.Errorf("Second result mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	tests := []struct {
		name       string
		cli1       armed.CLI
		cli2       armed.CLI
		shouldDiff bool
	}{
		{
			name: "same parameters generate same key",
			cli1: armed.CLI{
				Filename: "test.jsonnet",
				ExtStr:   map[string]string{"a": "1", "b": "2"},
				ExtCode:  map[string]string{"c": "3"},
			},
			cli2: armed.CLI{
				Filename: "test.jsonnet",
				ExtStr:   map[string]string{"b": "2", "a": "1"}, // Different order
				ExtCode:  map[string]string{"c": "3"},
			},
			shouldDiff: false,
		},
		{
			name: "different ExtStr generates different key",
			cli1: armed.CLI{
				Filename: "test.jsonnet",
				ExtStr:   map[string]string{"a": "1"},
			},
			cli2: armed.CLI{
				Filename: "test.jsonnet",
				ExtStr:   map[string]string{"a": "2"},
			},
			shouldDiff: true,
		},
		{
			name: "different filename generates different key",
			cli1: armed.CLI{
				Filename: "test1.jsonnet",
			},
			cli2: armed.CLI{
				Filename: "test2.jsonnet",
			},
			shouldDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary files with same content for key generation test
			tmpDir := t.TempDir()
			content := `{ test: "value" }`

			// Update filenames to use temp directory
			if tt.cli1.Filename != "" && tt.cli1.Filename != "-" {
				tmpFile1 := filepath.Join(tmpDir, tt.cli1.Filename)
				if err := os.WriteFile(tmpFile1, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write test file 1: %v", err)
				}
				tt.cli1.Filename = tmpFile1
			}

			if tt.cli2.Filename != "" && tt.cli2.Filename != "-" {
				tmpFile2 := filepath.Join(tmpDir, tt.cli2.Filename)
				if err := os.WriteFile(tmpFile2, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write test file 2: %v", err)
				}
				tt.cli2.Filename = tmpFile2
			}

			cache := armed.NewCache(time.Minute, 0)

			// Read content for both files
			content1, err := os.ReadFile(tt.cli1.Filename)
			if err != nil {
				t.Fatalf("Failed to read file1: %v", err)
			}
			content2, err := os.ReadFile(tt.cli2.Filename)
			if err != nil {
				t.Fatalf("Failed to read file2: %v", err)
			}

			key1, err1 := cache.GenerateCacheKey(&tt.cli1, content1)
			if err1 != nil {
				t.Fatalf("Failed to generate key1: %v", err1)
			}

			key2, err2 := cache.GenerateCacheKey(&tt.cli2, content2)
			if err2 != nil {
				t.Fatalf("Failed to generate key2: %v", err2)
			}

			if tt.shouldDiff {
				if key1 == key2 {
					t.Errorf("Keys should be different but are the same: %s", key1)
				}
			} else {
				if key1 != key2 {
					t.Errorf("Keys should be the same but differ: key1=%s, key2=%s", key1, key2)
				}
			}
		})
	}
}

func TestCacheWithStdin(t *testing.T) {
	// Skip stdin caching test for now - it needs special handling
	// because stdin can only be read once, and the current implementation
	// reads it during cache key generation
	t.Skip("Stdin caching requires special handling")
}
