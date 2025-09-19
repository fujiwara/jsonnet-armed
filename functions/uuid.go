package functions

import (
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/google/uuid"
)

// uuidV4Function generates a random UUID v4
func uuidV4Function(args []any) (any, error) {
	return uuid.New().String(), nil
}

// uuidV7Function generates a time-ordered UUID v7
func uuidV7Function(args []any) (any, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return u.String(), nil
}

var UuidFunctions = map[string]*jsonnet.NativeFunction{
	"uuid_v4": {
		Params: []ast.Identifier{},
		Func:   uuidV4Function,
	},
	"uuid_v7": {
		Params: []ast.Identifier{},
		Func:   uuidV7Function,
	},
}

func init() {
	initializeFunctionMap(UuidFunctions)
}
