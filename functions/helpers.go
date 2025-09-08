package functions

import (
	"github.com/google/go-jsonnet"
)

// initializeFunctionMap sets the Name field for all functions based on their map keys
func initializeFunctionMap(functionMap map[string]*jsonnet.NativeFunction) {
	for name, function := range functionMap {
		function.Name = name
	}
}
