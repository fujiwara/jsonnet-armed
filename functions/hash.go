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

// hashFileFunction creates a generic file hash function using the hash.Hash interface
func hashFileFunction(name string, newHasher func() hash.Hash) *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   name,
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("%s: filename must be a string", name)
			}
			
			file, err := os.Open(filename)
			if err != nil {
				return nil, fmt.Errorf("%s: failed to open file %s: %w", name, filename, err)
			}
			defer file.Close()
			
			hasher := newHasher()
			if _, err := io.Copy(hasher, file); err != nil {
				return nil, fmt.Errorf("%s: failed to read file %s: %w", name, filename, err)
			}
			
			return hex.EncodeToString(hasher.Sum(nil)), nil
		},
	}
}

var HashFunctions = []*jsonnet.NativeFunction{
	// String hash functions
	hashFunction("md5", func() hash.Hash { return md5.New() }),
	hashFunction("sha1", func() hash.Hash { return sha1.New() }),
	hashFunction("sha256", func() hash.Hash { return sha256.New() }),
	hashFunction("sha512", func() hash.Hash { return sha512.New() }),
	
	// File hash functions
	hashFileFunction("md5_file", func() hash.Hash { return md5.New() }),
	hashFileFunction("sha1_file", func() hash.Hash { return sha1.New() }),
	hashFileFunction("sha256_file", func() hash.Hash { return sha256.New() }),
	hashFileFunction("sha512_file", func() hash.Hash { return sha512.New() }),
}