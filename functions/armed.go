package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-jsonnet"
)

func GenerateAllFunctions(ctx context.Context) []*jsonnet.NativeFunction {
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
	for _, f := range GenerateExecFunctions(ctx) {
		all = append(all, f)
	}

	return all
}

// GenerateArmedLib returns the armed library as a string
func GenerateArmedLib(funcs []*jsonnet.NativeFunction) string {
	var lines []string
	lines = append(lines, "{")

	// Add all function definitions
	for _, f := range funcs {
		lines = append(lines, fmt.Sprintf("  %s: std.native('%s'),", f.Name, f.Name))
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}
