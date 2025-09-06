package functions

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

var (
	// Global context for exec functions
	execContextMu sync.RWMutex
	execContext   context.Context = context.Background()
	
	// DefaultExecTimeout is the default timeout for exec commands
	DefaultExecTimeout = 30 * time.Second
)

// SetExecContext sets the parent context for exec functions
// This allows CLI timeout to be propagated to exec commands
func SetExecContext(ctx context.Context) {
	execContextMu.Lock()
	defer execContextMu.Unlock()
	execContext = ctx
}

// ResetExecContext resets the exec context to Background
// This should be called when CLI execution completes
func ResetExecContext() {
	execContextMu.Lock()
	defer execContextMu.Unlock()
	execContext = context.Background()
}

// getExecContext returns the current exec context
func getExecContext() context.Context {
	execContextMu.RLock()
	defer execContextMu.RUnlock()
	return execContext
}

var ExecFunctions = map[string]*jsonnet.NativeFunction{
	"exec": {
		Params: []ast.Identifier{"command", "args"},
		Func: func(args []any) (any, error) {
			command, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("exec: command must be a string")
			}

			var cmdArgs []string
			if args[1] != nil {
				argsSlice, ok := args[1].([]any)
				if !ok {
					return nil, fmt.Errorf("exec: args must be an array")
				}
				cmdArgs = make([]string, len(argsSlice))
				for i, arg := range argsSlice {
					argStr, ok := arg.(string)
					if !ok {
						return nil, fmt.Errorf("exec: all arguments must be strings")
					}
					cmdArgs[i] = argStr
				}
			}

			return executeCommand(command, cmdArgs, nil)
		},
	},
	"exec_with_env": {
		Params: []ast.Identifier{"command", "args", "env_vars"},
		Func: func(args []any) (any, error) {
			command, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("exec_with_env: command must be a string")
			}

			var cmdArgs []string
			if args[1] != nil {
				argsSlice, ok := args[1].([]any)
				if !ok {
					return nil, fmt.Errorf("exec_with_env: args must be an array")
				}
				cmdArgs = make([]string, len(argsSlice))
				for i, arg := range argsSlice {
					argStr, ok := arg.(string)
					if !ok {
						return nil, fmt.Errorf("exec_with_env: all arguments must be strings")
					}
					cmdArgs[i] = argStr
				}
			}

			var envVars []string
			if args[2] != nil {
				envMap, ok := args[2].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("exec_with_env: env_vars must be an object")
				}
				for key, value := range envMap {
					valueStr, ok := value.(string)
					if !ok {
						return nil, fmt.Errorf("exec_with_env: environment variable values must be strings")
					}
					envVars = append(envVars, fmt.Sprintf("%s=%s", key, valueStr))
				}
			}

			return executeCommand(command, cmdArgs, envVars)
		},
	},
}

func executeCommand(command string, args []string, envVars []string) (map[string]any, error) {
	// Use parent context if available, otherwise use Background
	parentCtx := getExecContext()
	
	// Add timeout to the parent context
	ctx, cancel := context.WithTimeout(parentCtx, DefaultExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	// Set WaitDelay to ensure SIGKILL is sent after SIGTERM
	// This gives the process 5 seconds to gracefully terminate after SIGTERM
	cmd.WaitDelay = 5 * time.Second

	if envVars != nil {
		cmd.Env = append(os.Environ(), envVars...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0

	// Check for context cancellation/timeout first
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command execution timed out")
		} else if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("command execution was cancelled")
		}
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return map[string]any{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
	}, nil
}

func init() {
	initializeFunctionMap(ExecFunctions)
}