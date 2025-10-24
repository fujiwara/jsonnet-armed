//go:build !linux

package functions

import (
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var NetworkFunctions = map[string]*jsonnet.NativeFunction{
	"net_port_listening": {
		Params: []ast.Identifier{"protocol", "port"},
		Func: func(args []any) (any, error) {
			return nil, fmt.Errorf("net_port_listening is only supported on Linux")
		},
	},
}

func init() {
	initializeFunctionMap(NetworkFunctions)
}
