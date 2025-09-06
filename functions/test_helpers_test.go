package functions_test

import (
	"fmt"

	"github.com/fujiwara/jsonnet-armed/functions"
)

// Helper functions to get specific functions by name from maps
func getEnvFunction(name string) (func([]any) (any, error), error) {
	f, ok := functions.EnvFunctions[name]
	if !ok {
		return nil, fmt.Errorf("env function %s not found", name)
	}
	return f.Func, nil
}

func getBase64Function(name string) (func([]any) (any, error), error) {
	f, ok := functions.Base64Functions[name]
	if !ok {
		return nil, fmt.Errorf("base64 function %s not found", name)
	}
	return f.Func, nil
}

func getHashFunction(name string) (func([]any) (any, error), error) {
	f, ok := functions.HashFunctions[name]
	if !ok {
		return nil, fmt.Errorf("hash function %s not found", name)
	}
	return f.Func, nil
}

func getFileFunction(name string) (func([]any) (any, error), error) {
	f, ok := functions.FileFunctions[name]
	if !ok {
		return nil, fmt.Errorf("file function %s not found", name)
	}
	return f.Func, nil
}

func getTimeFunction(name string) (func([]any) (any, error), error) {
	f, ok := functions.TimeFunctions[name]
	if !ok {
		return nil, fmt.Errorf("time function %s not found", name)
	}
	return f.Func, nil
}
