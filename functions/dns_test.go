package functions

import (
	"testing"
)

func TestDnsLookup(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		recordType  string
		expectError bool
	}{
		{
			name:        "A record lookup for google.com",
			hostname:    "google.com",
			recordType:  "A",
			expectError: false,
		},
		{
			name:        "AAAA record lookup for google.com",
			hostname:    "google.com",
			recordType:  "AAAA",
			expectError: false,
		},
		{
			name:        "MX record lookup for google.com",
			hostname:    "google.com",
			recordType:  "MX",
			expectError: false,
		},
		{
			name:        "TXT record lookup for google.com",
			hostname:    "google.com",
			recordType:  "TXT",
			expectError: false,
		},
		{
			name:        "PTR record lookup for 8.8.8.8",
			hostname:    "8.8.8.8",
			recordType:  "PTR",
			expectError: false,
		},
		{
			name:        "NS record lookup for google.com",
			hostname:    "google.com",
			recordType:  "NS",
			expectError: false,
		},
		{
			name:        "case insensitive record type",
			hostname:    "google.com",
			recordType:  "a",
			expectError: false,
		},
		{
			name:        "unsupported record type",
			hostname:    "google.com",
			recordType:  "INVALID",
			expectError: true,
		},
		{
			name:        "HTTPS record lookup for cloudflare.com",
			hostname:    "cloudflare.com",
			recordType:  "HTTPS",
			expectError: false,
		},
		{
			name:        "SVCB record lookup for cloudflare.com",
			hostname:    "cloudflare.com",
			recordType:  "SVCB",
			expectError: false,
		},
		{
			name:        "non-existent domain",
			hostname:    "this-domain-does-not-exist-12345.com",
			recordType:  "A",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dnslookup(tt.hostname, tt.recordType)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate result structure
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Errorf("result is not a map[string]any")
				return
			}

			// Check required fields
			if hostname := resultMap["hostname"]; hostname != tt.hostname {
				t.Errorf("hostname mismatch: got %v, want %s", hostname, tt.hostname)
			}

			if recordType := resultMap["type"]; recordType != tt.recordType && recordType != "A" {
				// Allow "A" for case-insensitive test
				if !(tt.recordType == "a" && recordType == "A") {
					t.Errorf("type mismatch: got %v, want %s", recordType, tt.recordType)
				}
			}

			if success := resultMap["success"]; success != true {
				t.Errorf("success should be true, got %v", success)
			}

			records := resultMap["records"]
			if records == nil {
				t.Errorf("records field is missing")
				return
			}

			// Type-specific validations
			switch tt.recordType {
			case "A", "a":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("A records should be []any, got %T", records)
				} else if len(recordSlice) == 0 {
					t.Errorf("A records should not be empty")
				}

			case "AAAA":
				if _, ok := records.([]any); !ok {
					t.Errorf("AAAA records should be []any, got %T", records)
				}
				// IPv6 records may be empty for some domains

			case "MX":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("MX records should be []any, got %T", records)
				} else if len(recordSlice) > 0 {
					// Check first MX record structure
					if mxRecord, ok := recordSlice[0].(map[string]any); !ok {
						t.Errorf("MX record should be map[string]any, got %T", recordSlice[0])
					} else {
						if _, exists := mxRecord["priority"]; !exists {
							t.Errorf("MX record missing priority field")
						}
						if _, exists := mxRecord["hostname"]; !exists {
							t.Errorf("MX record missing hostname field")
						}
					}
				}

			case "TXT":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("TXT records should be []any, got %T", records)
				} else if len(recordSlice) == 0 {
					t.Errorf("TXT records should not be empty for google.com")
				}

			case "PTR":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("PTR records should be []any, got %T", records)
				} else if len(recordSlice) == 0 {
					t.Errorf("PTR records should not be empty for 8.8.8.8")
				}

			case "NS":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("NS records should be []any, got %T", records)
				} else if len(recordSlice) == 0 {
					t.Errorf("NS records should not be empty")
				}

			case "HTTPS", "SVCB":
				if recordSlice, ok := records.([]any); !ok {
					t.Errorf("HTTPS/SVCB records should be []any, got %T", records)
				} else {
					// HTTPS/SVCB records may be empty for some domains, so don't enforce non-empty
					// If not empty, validate structure
					if len(recordSlice) > 0 {
						if httpsRecord, ok := recordSlice[0].(map[string]any); ok {
							if _, exists := httpsRecord["priority"]; !exists {
								t.Errorf("HTTPS/SVCB record missing priority field")
							}
							if _, exists := httpsRecord["target"]; !exists {
								t.Errorf("HTTPS/SVCB record missing target field")
							}
							if _, exists := httpsRecord["params"]; !exists {
								t.Errorf("HTTPS/SVCB record missing params field")
							}
						}
					}
				}
			}
		})
	}
}

func TestDnsFunctions(t *testing.T) {
	// Test the actual Jsonnet function interface
	dnsFunc := DnsFunctions["dns_lookup"]
	if dnsFunc == nil {
		t.Fatal("dns_lookup function not found in DnsFunctions")
	}

	tests := []struct {
		name        string
		args        []any
		expectError bool
	}{
		{
			name:        "valid A record lookup",
			args:        []any{"google.com", "A"},
			expectError: false,
		},
		{
			name:        "invalid hostname type",
			args:        []any{123, "A"},
			expectError: true,
		},
		{
			name:        "invalid record type",
			args:        []any{"google.com", 123},
			expectError: true,
		},
		{
			name:        "unsupported record type",
			args:        []any{"google.com", "INVALID"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dnsFunc.Func(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Basic structure validation
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Errorf("result is not a map[string]any")
				return
			}

			if hostname := resultMap["hostname"]; hostname != tt.args[0] {
				t.Errorf("hostname mismatch: got %v, want %v", hostname, tt.args[0])
			}

			if recordType := resultMap["type"]; recordType != tt.args[1] {
				t.Errorf("type mismatch: got %v, want %v", recordType, tt.args[1])
			}
		})
	}
}

func TestDnsFunctionInitialization(t *testing.T) {
	// Test that function names are properly initialized
	for name, fn := range DnsFunctions {
		if fn.Name != name {
			t.Errorf("function %s has incorrect Name field: got %s, want %s", name, fn.Name, name)
		}
	}
}
