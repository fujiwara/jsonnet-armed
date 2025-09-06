package functions_test

import (
	"fmt"

	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-jsonnet"
)

// getFunctionByName returns a native function by its name from the given function list
func getFunctionByName(funcs []*jsonnet.NativeFunction, name string) (func([]any) (any, error), error) {
	for _, f := range funcs {
		if f.Name == name {
			return f.Func, nil
		}
	}
	return nil, fmt.Errorf("function %s not found", name)
}

// Helper functions to get specific functions by name
func getEnvFunction(name string) (func([]any) (any, error), error) {
	return getFunctionByName(functions.EnvFunctions, name)
}

func getBase64Function(name string) (func([]any) (any, error), error) {
	return getFunctionByName(functions.Base64Functions, name)
}

func getHashFunction(name string) (func([]any) (any, error), error) {
	return getFunctionByName(functions.HashFunctions, name)
}

func getFileFunction(name string) (func([]any) (any, error), error) {
	return getFunctionByName(functions.FileFunctions, name)
}

func getTimeFunction(name string) (func([]any) (any, error), error) {
	return getFunctionByName(functions.TimeFunctions, name)
}
