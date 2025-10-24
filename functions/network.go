//go:build linux

package functions

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// TCP_LISTEN state from Linux kernel
const TCP_LISTEN = "0A"

// checkPortListening checks if a port is listening by reading /proc/net/* files
func checkPortListening(port int, protocol string) (bool, error) {
	protocol = strings.ToLower(protocol)

	var procFile string
	switch protocol {
	case "tcp":
		procFile = "/proc/net/tcp"
	case "tcp6":
		procFile = "/proc/net/tcp6"
	case "udp":
		procFile = "/proc/net/udp"
	case "udp6":
		procFile = "/proc/net/udp6"
	default:
		return false, fmt.Errorf("unsupported protocol: %s (supported: tcp, tcp6, udp, udp6)", protocol)
	}

	file, err := os.Open(procFile)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", procFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header line
	if !scanner.Scan() {
		return false, fmt.Errorf("failed to read header from %s", procFile)
	}

	// Convert port to hex string (4 digits, uppercase)
	hexPort := fmt.Sprintf("%04X", port)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		// local_address is field[1], format: "IP:PORT"
		localAddr := fields[1]
		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}

		// Check if port matches
		if parts[1] == hexPort {
			// Check state (field[3])
			state := fields[3]
			// For TCP, check if state is LISTEN (0A)
			// For UDP, there's no LISTEN state, just check if port is bound
			if strings.HasPrefix(protocol, "tcp") {
				if state == TCP_LISTEN {
					return true, nil
				}
			} else {
				// UDP doesn't have connection state, if port is bound it's "listening"
				return true, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading %s: %w", procFile, err)
	}

	return false, nil
}

// parsePort parses a port number from various input types
func parsePort(portArg any) (int, error) {
	switch v := portArg.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		port, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("port must be a valid number")
		}
		return port, nil
	default:
		return 0, fmt.Errorf("port must be a number")
	}
}

var NetworkFunctions = map[string]*jsonnet.NativeFunction{
	"net_port_listening": {
		Params: []ast.Identifier{"protocol", "port"},
		Func: func(args []any) (any, error) {
			// Validate protocol argument
			protocol, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("net_port_listening: protocol must be a string")
			}

			// Validate and parse port argument
			port, err := parsePort(args[1])
			if err != nil {
				return nil, fmt.Errorf("net_port_listening: %w", err)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("net_port_listening: port must be between 1 and 65535, got %d", port)
			}

			return checkPortListening(port, protocol)
		},
	},
}

func init() {
	initializeFunctionMap(NetworkFunctions)
}
