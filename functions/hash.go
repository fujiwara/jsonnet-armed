package functions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var HashFunctions = []*jsonnet.NativeFunction{
	{
		Name:   "sha256",
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("sha256: data must be a string")
			}
			hasher := sha256.New()
			hasher.Write([]byte(data))
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	},
}