package functions

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var FileFunctions = []*jsonnet.NativeFunction{
	{
		Name:   "file_content",
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("file_content: filename must be a string")
			}

			file, err := os.Open(filename)
			if err != nil {
				return nil, fmt.Errorf("file_content: failed to open file %s: %w", filename, err)
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("file_content: failed to read file %s: %w", filename, err)
			}

			return string(content), nil
		},
	},
	{
		Name:   "file_stat",
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("file_stat: filename must be a string")
			}

			stat, err := os.Stat(filename)
			if err != nil {
				return nil, fmt.Errorf("file_stat: failed to stat file %s: %w", filename, err)
			}

			return map[string]any{
				"name":     stat.Name(),
				"size":     stat.Size(),
				"mode":     stat.Mode().String(),
				"mod_time": stat.ModTime().Unix(),
				"is_dir":   stat.IsDir(),
			}, nil
		},
	},
	{
		Name:   "file_exists",
		Params: []ast.Identifier{"filename"},
		Func: func(args []any) (any, error) {
			filename, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("file_exists: filename must be a string")
			}

			_, err := os.Stat(filename)
			return err == nil, nil
		},
	},
}
