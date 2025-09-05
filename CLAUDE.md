# jsonnet-armed Project Guidelines

## Project Overview
jsonnet-armed is a Jsonnet rendering tool with additional useful functions. It evaluates Jsonnet files and outputs JSON, with support for external variables.

## Code Style and Conventions

### Testing
- Use table-driven tests for all test cases
- Test packages should be named with `_test` suffix (e.g., `armed_test`)
- Use `github.com/google/go-cmp` for JSON comparison
- Compare JSON outputs by parsing them and using `cmp.Diff` for structural comparison
- Do not replace `os.Stdout` in tests; use `armed.SetOutput(io.Writer)` instead

### Test Structure Example
```go
tests := []struct {
    name        string
    jsonnet     string
    extStr      map[string]string
    extCode     map[string]string
    expected    string
    expectError bool
}{
    // test cases...
}
```

### JSON Comparison
```go
func compareJSON(t *testing.T, got, want string) {
    t.Helper()
    var gotJSON, wantJSON interface{}
    json.Unmarshal([]byte(got), &gotJSON)
    json.Unmarshal([]byte(want), &wantJSON)
    if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
        t.Errorf("JSON mismatch (-want +got):\n%s", diff)
    }
}
```

## Architecture Decisions

### Output Management
- Use `SetOutput(io.Writer)` function to control output destination
- Default output is `os.Stdout`
- Tests should capture output using `bytes.Buffer` with `SetOutput`

### External Variables
- Support both string variables (`--ext-str`) and code variables (`--ext-code`)
- String variables are passed as-is
- Code variables are evaluated as Jsonnet expressions

### Native Functions
- Environment functions: `env(name, default)` and `must_env(name)`
- Hash functions: `sha256(data)` returns SHA256 hash as hexadecimal string

## Development Commands

### Build
```bash
make
```

### Run Tests
```bash
go test -v ./...
```

### Install
```bash
make install
```

## Future Enhancements
- Native functions support (currently commented out in code)
- Additional Jsonnet extensions and utility functions