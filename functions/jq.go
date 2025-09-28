package functions

import (
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/itchyny/gojq"
)

var JQFunctions = map[string]*jsonnet.NativeFunction{
	"jq": {
		Params: []ast.Identifier{"query", "input"},
		Func: func(args []any) (any, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("jq: wrong number of arguments")
			}
			query, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("jq: argument must be a string")
			}
			input := args[1]

			q, err := gojq.Parse(query)
			if err != nil {
				return nil, fmt.Errorf("jq: failed to parse query: %v", err)
			}
			iter := q.Run(input)
			var results []any
			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
						break
					}
					return nil, fmt.Errorf("jq: failed to execute query: %v", err)
				}
				results = append(results, v)
			}
			switch len(results) {
			case 0:
				return nil, nil // No results
			case 1:
				return results[0], nil // Single result
			default:
				return results, nil // Multiple results
			}
		},
	},
}

func init() {
	initializeFunctionMap(JQFunctions)
}
