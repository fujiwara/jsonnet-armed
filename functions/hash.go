package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// hashFunction creates a generic hash function using the hash.Hash interface
func hashFunction(name string, newHasher func() hash.Hash) *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   name,
		Params: []ast.Identifier{"data"},
		Func: func(args []any) (any, error) {
			data, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("%s: data must be a string", name)
			}
			hasher := newHasher()
			hasher.Write([]byte(data))
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	}
}

var HashFunctions = []*jsonnet.NativeFunction{
	hashFunction("md5", func() hash.Hash { return md5.New() }),
	hashFunction("sha1", func() hash.Hash { return sha1.New() }),
	hashFunction("sha256", func() hash.Hash { return sha256.New() }),
	hashFunction("sha512", func() hash.Hash { return sha512.New() }),
}