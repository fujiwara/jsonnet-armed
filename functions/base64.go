package functions

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var Base64Functions = map[string]*jsonnet.NativeFunction{
	"base64": {
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("base64: data must be a string")
			}
			return base64.StdEncoding.EncodeToString([]byte(data)), nil
		},
	},
	"base64url": {
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("base64url: data must be a string")
			}
			return base64.URLEncoding.EncodeToString([]byte(data)), nil
		},
	},
}

func init() {
	initializeFunctionMap(Base64Functions)
}
