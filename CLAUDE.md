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
- When adding new native functions, add both unit tests (in `functions/*_test.go`) and integration tests (in `integration_test.go`)

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
- Environment functions: `env(name, default)`, `must_env(name)`, and `env_parse(content)`
  - `env_parse` parses .env format strings and returns a map[string]any for JSON compatibility
- Hash functions: `md5(data)`, `sha1(data)`, `sha256(data)`, `sha512(data)` return hash as hexadecimal string
- File hash functions: `md5_file(filename)`, `sha1_file(filename)`, `sha256_file(filename)`, `sha512_file(filename)` return file content hash as hexadecimal string
- File functions: `file_content(filename)` returns file content as string, `file_stat(filename)` returns file metadata object

### Native Function Implementation Notes
- All native functions must return JSON-compatible types (`any`/`interface{}`)
- When returning maps, use `map[string]any` instead of `map[string]string` for Jsonnet compatibility
- External dependencies (like `github.com/hashicorp/go-envparse`) should be added via `go get`
- Always add test coverage:
  - Unit tests in `functions/<function>_test.go` for detailed testing
  - At least one integration test case in `integration_test.go` to verify end-to-end functionality
  - If function reads files, create test fixtures in `testdata/` directory

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