package functions

import (
	"testing"
)

func TestParseCertificate(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		checkFields func(t *testing.T, result any)
	}{
		{
			name:        "Valid RSA certificate",
			filename:    "../testdata/test-rsa.crt",
			expectError: false,
			checkFields: func(t *testing.T, result any) {
				certMap, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not map[string]any")
				}

				// Check required fields exist
				requiredFields := []string{
					"fingerprint_sha1",
					"fingerprint_sha256",
					"public_key_fingerprint_sha256",
					"subject",
					"issuer",
					"serial_number",
					"not_before",
					"not_after",
					"not_before_unix",
					"not_after_unix",
					"dns_names",
					"ip_addresses",
					"is_ca",
					"version",
					"signature_algorithm",
					"public_key_algorithm",
				}

				for _, field := range requiredFields {
					if _, exists := certMap[field]; !exists {
						t.Errorf("missing required field: %s", field)
					}
				}

				// Check subject
				subject, ok := certMap["subject"].(map[string]any)
				if !ok {
					t.Fatal("subject is not map[string]any")
				}
				if subject["common_name"] != "test.example.com" {
					t.Errorf("unexpected common_name: %v", subject["common_name"])
				}

				// Check DNS names
				dnsNames, ok := certMap["dns_names"].([]any)
				if !ok {
					t.Fatal("dns_names is not []any")
				}
				if len(dnsNames) == 0 {
					t.Error("dns_names should not be empty")
				}

				// Check fingerprint format (XX:XX:XX:...)
				fpSha256, ok := certMap["fingerprint_sha256"].(string)
				if !ok {
					t.Fatal("fingerprint_sha256 is not string")
				}
				if len(fpSha256) != 95 { // 32 bytes * 2 hex + 31 colons
					t.Errorf("unexpected fingerprint_sha256 length: %d", len(fpSha256))
				}

				// Check public key algorithm
				pubKeyAlg, ok := certMap["public_key_algorithm"].(string)
				if !ok {
					t.Fatal("public_key_algorithm is not string")
				}
				if pubKeyAlg != "RSA" {
					t.Errorf("unexpected public_key_algorithm: %v", pubKeyAlg)
				}
			},
		},
		{
			name:        "Valid ECDSA certificate",
			filename:    "../testdata/test-ecdsa.crt",
			expectError: false,
			checkFields: func(t *testing.T, result any) {
				certMap, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not map[string]any")
				}

				// Check subject
				subject, ok := certMap["subject"].(map[string]any)
				if !ok {
					t.Fatal("subject is not map[string]any")
				}
				if subject["common_name"] != "ecdsa.example.com" {
					t.Errorf("unexpected common_name: %v", subject["common_name"])
				}

				// Check public key algorithm
				pubKeyAlg, ok := certMap["public_key_algorithm"].(string)
				if !ok {
					t.Fatal("public_key_algorithm is not string")
				}
				if pubKeyAlg != "ECDSA" {
					t.Errorf("unexpected public_key_algorithm: %v", pubKeyAlg)
				}
			},
		},
		{
			name:        "Non-existent file",
			filename:    "../testdata/nonexistent.crt",
			expectError: true,
		},
		{
			name:        "Invalid certificate file",
			filename:    "../testdata/test.env",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCertificate(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFields != nil {
				tt.checkFields(t, result)
			}
		})
	}
}

func TestParsePrivateKey(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		checkFields func(t *testing.T, result any)
	}{
		{
			name:        "Valid RSA private key",
			filename:    "../testdata/test-rsa.key",
			expectError: false,
			checkFields: func(t *testing.T, result any) {
				keyMap, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not map[string]any")
				}

				// Check key type
				keyType, ok := keyMap["key_type"].(string)
				if !ok || keyType != "RSA" {
					t.Errorf("unexpected key_type: %v", keyMap["key_type"])
				}

				// Check key size
				keySize, ok := keyMap["key_size"].(int)
				if !ok || keySize != 2048 {
					t.Errorf("unexpected key_size: %v", keyMap["key_size"])
				}

				// Check public key fingerprint
				fpSha256, ok := keyMap["public_key_fingerprint_sha256"].(string)
				if !ok {
					t.Fatal("public_key_fingerprint_sha256 is not string")
				}
				if len(fpSha256) != 95 { // 32 bytes * 2 hex + 31 colons
					t.Errorf("unexpected fingerprint length: %d", len(fpSha256))
				}

				// Check public key PEM
				pubKeyPEM, ok := keyMap["public_key_pem"].(string)
				if !ok {
					t.Fatal("public_key_pem is not string")
				}
				if len(pubKeyPEM) == 0 {
					t.Error("public_key_pem should not be empty")
				}
			},
		},
		{
			name:        "Valid ECDSA private key",
			filename:    "../testdata/test-ecdsa.key",
			expectError: false,
			checkFields: func(t *testing.T, result any) {
				keyMap, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not map[string]any")
				}

				// Check key type
				keyType, ok := keyMap["key_type"].(string)
				if !ok || keyType != "ECDSA" {
					t.Errorf("unexpected key_type: %v", keyMap["key_type"])
				}

				// Check curve
				curve, ok := keyMap["curve"].(string)
				if !ok || curve != "P-256" {
					t.Errorf("unexpected curve: %v", keyMap["curve"])
				}

				// Check public key fingerprint exists
				_, ok = keyMap["public_key_fingerprint_sha256"].(string)
				if !ok {
					t.Error("public_key_fingerprint_sha256 should exist")
				}
			},
		},
		{
			name:        "Non-existent file",
			filename:    "../testdata/nonexistent.key",
			expectError: true,
		},
		{
			name:        "Invalid key file",
			filename:    "../testdata/test.env",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePrivateKey(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFields != nil {
				tt.checkFields(t, result)
			}
		})
	}
}

func TestCertificateKeyPairMatch(t *testing.T) {
	// Test that certificate and private key have matching public key fingerprints
	certResult, err := parseCertificate("../testdata/test-rsa.crt")
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	keyResult, err := parsePrivateKey("../testdata/test-rsa.key")
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	certMap := certResult.(map[string]any)
	keyMap := keyResult.(map[string]any)

	certFP := certMap["public_key_fingerprint_sha256"].(string)
	keyFP := keyMap["public_key_fingerprint_sha256"].(string)

	if certFP != keyFP {
		t.Errorf("certificate and key public key fingerprints do not match:\ncert: %s\nkey:  %s", certFP, keyFP)
	}
}

func TestX509Functions(t *testing.T) {
	// Test x509_certificate function
	certFunc := X509Functions["x509_certificate"]
	if certFunc == nil {
		t.Fatal("x509_certificate function not found")
	}

	certResult, err := certFunc.Func([]any{"../testdata/test-rsa.crt"})
	if err != nil {
		t.Fatalf("x509_certificate failed: %v", err)
	}

	certMap, ok := certResult.(map[string]any)
	if !ok {
		t.Fatal("x509_certificate result is not map[string]any")
	}

	if _, exists := certMap["fingerprint_sha256"]; !exists {
		t.Error("fingerprint_sha256 field missing")
	}

	// Test x509_private_key function
	keyFunc := X509Functions["x509_private_key"]
	if keyFunc == nil {
		t.Fatal("x509_private_key function not found")
	}

	keyResult, err := keyFunc.Func([]any{"../testdata/test-rsa.key"})
	if err != nil {
		t.Fatalf("x509_private_key failed: %v", err)
	}

	keyMap, ok := keyResult.(map[string]any)
	if !ok {
		t.Fatal("x509_private_key result is not map[string]any")
	}

	if _, exists := keyMap["public_key_fingerprint_sha256"]; !exists {
		t.Error("public_key_fingerprint_sha256 field missing")
	}

	// Test invalid argument type
	_, err = certFunc.Func([]any{123})
	if err == nil {
		t.Error("expected error for invalid argument type")
	}
}

func TestX509FunctionInitialization(t *testing.T) {
	// Test that function names are properly initialized
	for name, fn := range X509Functions {
		if fn.Name != name {
			t.Errorf("function %s has incorrect Name field: got %s, want %s", name, fn.Name, name)
		}
	}
}
