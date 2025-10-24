//go:build linux

package functions

import (
	"net"
	"testing"
	"time"
)

func TestCheckPortListening(t *testing.T) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	testPort := listener.Addr().(*net.TCPAddr).Port

	// Start a test UDP server
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve UDP address: %v", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("failed to start UDP server: %v", err)
	}
	defer udpConn.Close()

	testUDPPort := udpConn.LocalAddr().(*net.UDPAddr).Port

	// Wait a bit for the servers to be fully ready
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name        string
		port        int
		protocol    string
		expected    bool
		expectError bool
	}{
		{
			name:        "TCP port listening",
			port:        testPort,
			protocol:    "tcp",
			expected:    true,
			expectError: false,
		},
		{
			name:        "TCP6 same port should be listening (dual stack)",
			port:        testPort,
			protocol:    "tcp6",
			expected:    false, // May or may not be listening on IPv6
			expectError: false,
		},
		{
			name:        "UDP port listening",
			port:        testUDPPort,
			protocol:    "udp",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Port not listening",
			port:        65534, // Very unlikely to be in use
			protocol:    "tcp",
			expected:    false,
			expectError: false,
		},
		{
			name:        "UDP port not listening",
			port:        65533,
			protocol:    "udp",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Invalid protocol",
			port:        80,
			protocol:    "invalid",
			expected:    false,
			expectError: true,
		},
		{
			name:        "Case insensitive protocol - TCP",
			port:        testPort,
			protocol:    "TCP",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Case insensitive protocol - Tcp",
			port:        testPort,
			protocol:    "Tcp",
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checkPortListening(tt.port, tt.protocol)

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

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    int
		expectError bool
	}{
		{
			name:        "float64 port",
			input:       float64(8080),
			expected:    8080,
			expectError: false,
		},
		{
			name:        "int port",
			input:       8080,
			expected:    8080,
			expectError: false,
		},
		{
			name:        "string port",
			input:       "8080",
			expected:    8080,
			expectError: false,
		},
		{
			name:        "invalid string port",
			input:       "abc",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid type",
			input:       []int{8080},
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePort(tt.input)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestNetworkFunctions(t *testing.T) {
	// Start a test TCP server for function interface tests
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	testPort := listener.Addr().(*net.TCPAddr).Port
	time.Sleep(100 * time.Millisecond)

	netFunc := NetworkFunctions["net_port_listening"]
	if netFunc == nil {
		t.Fatal("net_port_listening function not found in NetworkFunctions")
	}

	tests := []struct {
		name        string
		args        []any
		expected    bool
		expectError bool
	}{
		{
			name:        "valid listening port",
			args:        []any{"tcp", float64(testPort)},
			expected:    true,
			expectError: false,
		},
		{
			name:        "valid non-listening port",
			args:        []any{"tcp", float64(65534)},
			expected:    false,
			expectError: false,
		},
		{
			name:        "invalid port type - string",
			args:        []any{"tcp", "8080"},
			expected:    false,
			expectError: false, // parsePort should handle string conversion
		},
		{
			name:        "invalid port range - too low",
			args:        []any{"tcp", float64(0)},
			expected:    false,
			expectError: true,
		},
		{
			name:        "invalid port range - too high",
			args:        []any{"tcp", float64(65536)},
			expected:    false,
			expectError: true,
		},
		{
			name:        "invalid protocol type",
			args:        []any{123, float64(80)},
			expected:    false,
			expectError: true,
		},
		{
			name:        "unsupported protocol",
			args:        []any{"sctp", float64(80)},
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := netFunc.Func(tt.args)

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

			boolResult, ok := result.(bool)
			if !ok {
				t.Errorf("result is not a bool, got %T", result)
				return
			}

			if boolResult != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, boolResult)
			}
		})
	}
}

func TestNetworkFunctionInitialization(t *testing.T) {
	// Test that function names are properly initialized
	for name, fn := range NetworkFunctions {
		if fn.Name != name {
			t.Errorf("function %s has incorrect Name field: got %s, want %s", name, fn.Name, name)
		}
	}
}
