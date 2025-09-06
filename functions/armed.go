package functions

import (
	"strings"

	"github.com/google/go-jsonnet"
)

// AllFunctions returns all native functions in one slice
func AllFunctions() []*jsonnet.NativeFunction {
	var all []*jsonnet.NativeFunction

	// Add functions from maps
	for _, f := range EnvFunctions {
		all = append(all, f)
	}
	for _, f := range HashFunctions {
		all = append(all, f)
	}
	for _, f := range FileFunctions {
		all = append(all, f)
	}
	for _, f := range Base64Functions {
		all = append(all, f)
	}
	for _, f := range TimeFunctions {
		all = append(all, f)
	}

	return all
}

// GenerateArmedLib returns the armed library as a string
func GenerateArmedLib() string {
	var lines []string
	lines = append(lines, "{")

	// Add all function definitions
	for _, f := range AllFunctions() {
		lines = append(lines, "  "+f.Name+": std.native('"+f.Name+"'),")
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}
