# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

jsonnet-armed is a Go CLI tool that extends the standard Jsonnet evaluator with native functions useful for infrastructure/DevOps configuration (env vars, hashing, HTTP, DNS, exec, regex, jq, UUID, x509, etc.). It evaluates Jsonnet files and outputs JSON.

## Development Commands

```bash
make                          # Build binary
make test                     # Run all tests with -race
go test -v ./...              # Run all tests (verbose)
go test -v ./functions        # Run only unit tests for native functions
go test -v -run TestName      # Run a single test by name
go test -v -run TestName ./functions  # Run a single function test
make install                  # Install binary
go fmt ./...                  # Format code (run before every commit)
go mod tidy                   # Clean dependencies (run after adding deps)
```

## Architecture

### Package Structure

- **`armed` (root package)**: CLI lifecycle, Jsonnet VM orchestration, caching, output handling. Key types: `CLI` (kong-based), `Cache`, `ArmedImporter`.
- **`functions` package**: All native function implementations. No knowledge of the `armed` package (one-way dependency).
- **`cmd/jsonnet-armed`**: Binary entry point with signal handling.

### Execution Flow

`cmd/main.go` → `armed.Run()` → kong CLI parsing → `cli.run()` → `cli.processRequest()` → `cli.evaluate()` → creates `jsonnet.VM`, registers all native functions via `functions.GenerateAllFunctions(ctx)`, sets `ArmedImporter` (provides virtual `armed.libsonnet`), evaluates Jsonnet.

### Native Function Registration Pattern

Functions follow two patterns:

**Static maps** (context-independent) — most functions use this:
```go
// In functions/<category>.go
var CategoryFunctions = map[string]*jsonnet.NativeFunction{
    "func_name": {
        Params: []ast.Identifier{"param1", "param2"},
        Func: func(args []any) (any, error) { ... },
    },
}
func init() { initializeFunctionMap(CategoryFunctions) }  // sets Name from map key
```

**Generator functions** (context-dependent) — used by `exec` and `http`:
```go
func GenerateExecFunctions(ctx context.Context) map[string]*jsonnet.NativeFunction { ... }
```

All function maps are aggregated in `functions/armed.go:GenerateAllFunctions()`. When adding a new category, add its map iteration there.

### ArmedImporter

Custom Jsonnet importer that intercepts `import 'armed.libsonnet'` and dynamically generates a Jsonnet object mapping all function names to `std.native()` calls. Users can use either `std.native("func")` or `(import 'armed.libsonnet').func`.

## Adding New Native Functions

1. Create `functions/<category>.go` with exported map (e.g., `CategoryFunctions`)
2. Use `init()` to call `initializeFunctionMap()`
3. Register in `functions/armed.go:GenerateAllFunctions()`
4. All functions must return `(any, error)` with JSON-compatible types (`map[string]any`, `[]any`, not typed maps/slices)
5. Add unit tests in `functions/<category>_test.go` (package `functions_test`, table-driven)
6. Add integration test case in `integration_test.go` (package `armed_test`)
7. Create test fixtures in `testdata/` if function reads files

## Testing Conventions

- Table-driven tests with `[]struct{ name, args/jsonnet, expected, expectError }`
- Use `github.com/google/go-cmp/cmp.Diff` for JSON structural comparison
- Use `cli.SetWriter(&buf)` to capture output in tests — never replace `os.Stdout`
- For non-deterministic outputs (UUIDs, timestamps): test format validity with regex, use validation placeholders like `<valid_uuid_v4>` in integration tests
- Unit test helpers in `functions/test_helpers_test.go` provide `getEnvFunction()`, `getHashFunction()`, etc.

## Pre-commit Checklist

1. `go fmt ./...`
2. `go mod tidy`
3. `go test -v ./functions` (unit tests)
4. `go test -v` (integration tests)
