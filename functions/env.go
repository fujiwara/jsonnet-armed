package functions

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/hashicorp/go-envparse"
)

var EnvFunctions = map[string]*jsonnet.NativeFunction{
	"env": {
		Params: []ast.Identifier{"name", "default"},
		Func: func(args []any) (any, error) {
			key, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("env: name must be a string")
			}
			if v := os.Getenv(key); v != "" {
				return v, nil
			}
			return args[1], nil
		},
	},
	"must_env": {
		Params: []ast.Identifier{"name"},
		Func: func(args []any) (any, error) {
			key, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("must_env: name must be a string")
			}
			if v, ok := os.LookupEnv(key); ok {
				return v, nil
			}
			return nil, fmt.Errorf("must_env: %s is not set", key)
		},
	},
	"env_parse": {
		Params: []ast.Identifier{"content"},
		Func: func(args []any) (any, error) {
			content, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("env_parse: content must be a string")
			}
			envMap, err := envparse.Parse(strings.NewReader(content))
			if err != nil {
				return nil, fmt.Errorf("env_parse: failed to parse: %w", err)
			}
			// Convert map[string]string to map[string]any for JSON compatibility
			result := make(map[string]any)
			for k, v := range envMap {
				result[k] = v
			}
			return result, nil
		},
	},
}

func init() {
	initializeFunctionMap(EnvFunctions)
}
