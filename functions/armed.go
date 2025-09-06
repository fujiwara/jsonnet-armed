package functions

import (
	"strings"

	"github.com/google/go-jsonnet"
)

// AllFunctions returns all native functions in one slice
func AllFunctions() []*jsonnet.NativeFunction {
	var all []*jsonnet.NativeFunction
	all = append(all, EnvFunctions...)
	all = append(all, HashFunctions...)
	all = append(all, FileFunctions...)
	all = append(all, Base64Functions...)
	all = append(all, TimeFunctions...)
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
