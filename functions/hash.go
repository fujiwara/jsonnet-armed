package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var HashFunctions = []*jsonnet.NativeFunction{
	{
		Name:   "md5",
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("md5: data must be a string")
			}
			hasher := md5.New()
			hasher.Write([]byte(data))
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	},
	{
		Name:   "sha1",
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("sha1: data must be a string")
			}
			hasher := sha1.New()
			hasher.Write([]byte(data))
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	},
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
	{
		Name:   "sha512",
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("sha512: data must be a string")
			}
			hasher := sha512.New()
			hasher.Write([]byte(data))
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	},
}