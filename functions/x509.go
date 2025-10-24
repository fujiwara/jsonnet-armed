package functions

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// formatFingerprint formats a byte slice as a colon-separated hex string
func formatFingerprint(data []byte) string {
	hex := hex.EncodeToString(data)
	var parts []string
	for i := 0; i < len(hex); i += 2 {
		parts = append(parts, strings.ToUpper(hex[i:i+2]))
	}
	return strings.Join(parts, ":")
}

// publicKeyFingerprint calculates the SHA256 fingerprint of a public key
func publicKeyFingerprint(pubKey crypto.PublicKey) (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	hash := sha256.Sum256(pubKeyBytes)
	return formatFingerprint(hash[:]), nil
}

// parseCertificate parses an X.509 certificate file and returns detailed information
func parseCertificate(filename string) (any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Calculate fingerprints
	sha1Hash := sha1.Sum(cert.Raw)
	sha256Hash := sha256.Sum256(cert.Raw)

	pubKeyFP, err := publicKeyFingerprint(cert.PublicKey)
	if err != nil {
		return nil, err
	}

	// Extract subject information
	subject := map[string]any{
		"common_name":         cert.Subject.CommonName,
		"organization":        convertToAny(cert.Subject.Organization),
		"organizational_unit": convertToAny(cert.Subject.OrganizationalUnit),
		"country":             convertToAny(cert.Subject.Country),
		"province":            convertToAny(cert.Subject.Province),
		"locality":            convertToAny(cert.Subject.Locality),
	}

	// Extract issuer information
	issuer := map[string]any{
		"common_name":         cert.Issuer.CommonName,
		"organization":        convertToAny(cert.Issuer.Organization),
		"organizational_unit": convertToAny(cert.Issuer.OrganizationalUnit),
		"country":             convertToAny(cert.Issuer.Country),
		"province":            convertToAny(cert.Issuer.Province),
		"locality":            convertToAny(cert.Issuer.Locality),
	}

	// Convert DNS names, IP addresses, email addresses
	dnsNames := convertToAny(cert.DNSNames)
	emailAddresses := convertToAny(cert.EmailAddresses)

	var ipAddresses []any
	for _, ip := range cert.IPAddresses {
		ipAddresses = append(ipAddresses, ip.String())
	}

	// Key usage
	var keyUsage []any
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		keyUsage = append(keyUsage, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		keyUsage = append(keyUsage, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		keyUsage = append(keyUsage, "Data Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		keyUsage = append(keyUsage, "Key Agreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		keyUsage = append(keyUsage, "Certificate Sign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		keyUsage = append(keyUsage, "CRL Sign")
	}

	// Extended key usage
	var extKeyUsage []any
	for _, eku := range cert.ExtKeyUsage {
		switch eku {
		case x509.ExtKeyUsageServerAuth:
			extKeyUsage = append(extKeyUsage, "Server Auth")
		case x509.ExtKeyUsageClientAuth:
			extKeyUsage = append(extKeyUsage, "Client Auth")
		case x509.ExtKeyUsageCodeSigning:
			extKeyUsage = append(extKeyUsage, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			extKeyUsage = append(extKeyUsage, "Email Protection")
		case x509.ExtKeyUsageTimeStamping:
			extKeyUsage = append(extKeyUsage, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			extKeyUsage = append(extKeyUsage, "OCSP Signing")
		}
	}

	return map[string]any{
		"fingerprint_sha1":              formatFingerprint(sha1Hash[:]),
		"fingerprint_sha256":            formatFingerprint(sha256Hash[:]),
		"public_key_fingerprint_sha256": pubKeyFP,
		"subject":                       subject,
		"issuer":                        issuer,
		"serial_number":                 cert.SerialNumber.String(),
		"not_before":                    cert.NotBefore.Format("2006-01-02T15:04:05Z07:00"),
		"not_after":                     cert.NotAfter.Format("2006-01-02T15:04:05Z07:00"),
		"not_before_unix":               cert.NotBefore.Unix(),
		"not_after_unix":                cert.NotAfter.Unix(),
		"dns_names":                     dnsNames,
		"ip_addresses":                  ipAddresses,
		"email_addresses":               emailAddresses,
		"is_ca":                         cert.IsCA,
		"version":                       cert.Version,
		"signature_algorithm":           cert.SignatureAlgorithm.String(),
		"public_key_algorithm":          cert.PublicKeyAlgorithm.String(),
		"key_usage":                     keyUsage,
		"ext_key_usage":                 extKeyUsage,
	}, nil
}

// parsePrivateKey parses a private key file and returns information (without exposing the key)
func parsePrivateKey(filename string) (any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	var privKey crypto.PrivateKey
	var parseErr error

	// Decode all PEM blocks and find the private key
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		data = rest

		// Try parsing as different key types
		switch block.Type {
		case "RSA PRIVATE KEY":
			privKey, parseErr = x509.ParsePKCS1PrivateKey(block.Bytes)
		case "EC PRIVATE KEY":
			privKey, parseErr = x509.ParseECPrivateKey(block.Bytes)
		case "PRIVATE KEY":
			privKey, parseErr = x509.ParsePKCS8PrivateKey(block.Bytes)
		case "EC PARAMETERS":
			// Skip EC PARAMETERS blocks
			continue
		default:
			continue
		}

		// If parsing succeeded, we're done
		if parseErr == nil && privKey != nil {
			break
		}
	}

	if privKey == nil {
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", parseErr)
		}
		return nil, fmt.Errorf("no private key found in file")
	}

	result := map[string]any{}

	// Extract public key and calculate fingerprint
	var pubKey crypto.PublicKey
	var keyType string
	var keySize int
	var curve string

	switch key := privKey.(type) {
	case *rsa.PrivateKey:
		keyType = "RSA"
		keySize = key.N.BitLen()
		pubKey = &key.PublicKey
	case *ecdsa.PrivateKey:
		keyType = "ECDSA"
		curve = key.Curve.Params().Name
		pubKey = &key.PublicKey
	case ed25519.PrivateKey:
		keyType = "Ed25519"
		pubKey = key.Public()
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", privKey)
	}

	result["key_type"] = keyType
	if keySize > 0 {
		result["key_size"] = keySize
	}
	if curve != "" {
		result["curve"] = curve
	}

	// Calculate public key fingerprint
	pubKeyFP, err := publicKeyFingerprint(pubKey)
	if err != nil {
		return nil, err
	}
	result["public_key_fingerprint_sha256"] = pubKeyFP

	// Marshal public key to PEM format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	result["public_key_pem"] = string(pubKeyPEM)

	return result, nil
}

// convertToAny converts a string slice to []any for JSON compatibility
func convertToAny(s []string) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

var X509Functions = map[string]*jsonnet.NativeFunction{
	"x509_certificate": {
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("x509_certificate: filename must be a string")
			}
			return parseCertificate(filename)
		},
	},
	"x509_private_key": {
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("x509_private_key: filename must be a string")
			}
			return parsePrivateKey(filename)
		},
	},
}

func init() {
	initializeFunctionMap(X509Functions)
}
