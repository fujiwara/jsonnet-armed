package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

// hashFunction creates a generic hash function using the hash.Hash interface
func hashFunction(newHasher func() hash.Hash) func([]any) (any, error) {
	return func(args []any) (any, error) {
		data, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("data must be a string")
		}
		hasher := newHasher()
		hasher.Write([]byte(data))
		return hex.EncodeToString(hasher.Sum(nil)), nil
	}
}

// hashFileFunction creates a generic file hash function using the hash.Hash interface
func hashFileFunction(newHasher func() hash.Hash) func([]any) (any, error) {
	return func(args []any) (any, error) {
		filename, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("filename must be a string")
		}

		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
		}
		defer file.Close()

		hasher := newHasher()
		if _, err := io.Copy(hasher, file); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		return hex.EncodeToString(hasher.Sum(nil)), nil
	}
}

var HashFunctions = map[string]*jsonnet.NativeFunction{
	// String hash functions
	"md5": {
		Params: []ast.Identifier{"data"},
		Func:   hashFunction(func() hash.Hash { return md5.New() }),
	},
	"sha1": {
		Params: []ast.Identifier{"data"},
		Func:   hashFunction(func() hash.Hash { return sha1.New() }),
	},
	"sha256": {
		Params: []ast.Identifier{"data"},
		Func:   hashFunction(func() hash.Hash { return sha256.New() }),
	},
	"sha512": {
		Params: []ast.Identifier{"data"},
		Func:   hashFunction(func() hash.Hash { return sha512.New() }),
	},

	// File hash functions
	"md5_file": {
		Params: []ast.Identifier{"filename"},
		Func:   hashFileFunction(func() hash.Hash { return md5.New() }),
	},
	"sha1_file": {
		Params: []ast.Identifier{"filename"},
		Func:   hashFileFunction(func() hash.Hash { return sha1.New() }),
	},
	"sha256_file": {
		Params: []ast.Identifier{"filename"},
		Func:   hashFileFunction(func() hash.Hash { return sha256.New() }),
	},
	"sha512_file": {
		Params: []ast.Identifier{"filename"},
		Func:   hashFileFunction(func() hash.Hash { return sha512.New() }),
	},
}

func init() {
	initializeFunctionMap(HashFunctions)
}
