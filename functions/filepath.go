package functions

import (
	"fmt"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var PathFunctions = map[string]*jsonnet.NativeFunction{
	"basename": {
		Params: []ast.Identifier{"path"},
		Func: func(args []any) (any, error) {
			path, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("basename: path must be a string")
			}
			return filepath.Base(path), nil
		},
	},
	"dirname": {
		Params: []ast.Identifier{"path"},
		Func: func(args []any) (any, error) {
			path, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("dirname: path must be a string")
			}
			return filepath.Dir(path), nil
		},
	},
	"extname": {
		Params: []ast.Identifier{"path"},
		Func: func(args []any) (any, error) {
			path, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("extname: path must be a string")
			}
			return filepath.Ext(path), nil
		},
	},
	"path_join": {
		Params: []ast.Identifier{"elements"},
		Func: func(args []any) (any, error) {
			elements, ok := args[0].([]any)
			if !ok {
				return nil, fmt.Errorf("path_join: elements must be an array")
			}
			parts := make([]string, len(elements))
			for i, e := range elements {
				s, ok := e.(string)
				if !ok {
					return nil, fmt.Errorf("path_join: element at index %d must be a string", i)
				}
				parts[i] = s
			}
			return filepath.Join(parts...), nil
		},
	},
}

func init() {
	initializeFunctionMap(PathFunctions)
}
