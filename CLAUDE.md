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

## Development Workflow

### Go Module Management
- Always run `go mod tidy` after adding or updating dependencies
- This ensures go.mod and go.sum are properly synchronized
- Example workflow:
  ```bash
  go get github.com/some/package
  go mod tidy
  ```

### Code Formatting
- Always run `go fmt ./...` before committing code
- This ensures consistent code formatting across the project
- Make it a habit to format before every commit

### Git Commit Best Practices
- Use `git add <file>` to add specific files instead of `git add -A`
- This prevents accidentally committing unintended changes
- Review each file before adding:
  ```bash
  git status
  git add functions/new_function.go
  git add functions/new_function_test.go
  git add go.mod go.sum
  git commit -m "Add new function"
  ```

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

## Native Function Development Guidelines

### Function Planning and Research
- Before implementing new functions, research existing Jsonnet standard library functions to avoid duplication
- Use reliable, well-maintained libraries (e.g., `github.com/google/uuid` for UUID generation)
- Prioritize functions that fill gaps in infrastructure/DevOps configuration use cases
- Consider security implications: use proven cryptographic libraries rather than custom implementations

### Function Categories and Naming
- Follow consistent naming patterns: `<category>_<operation>` (e.g., `uuid_v4`, `regex_match`)
- Group related functions in single files (e.g., `uuid.go` for all UUID functions)
- Export function maps with descriptive names (e.g., `UuidFunctions`, `RegexpFunctions`)

### Function Implementation Patterns
- All native functions should accept `[]any` parameters and return `(any, error)`
- Validate input types early and return descriptive errors
- Return JSON-compatible types: use `[]any` instead of `[]string` for arrays
- Handle edge cases gracefully (e.g., return empty arrays instead of nil for no matches)

### Function Registration
- Register function maps in `functions/armed.go` within `GenerateAllFunctions()`
- Use `initializeFunctionMap()` helper in each function file's `init()` function
- Ensure functions are available both as native functions and in the armed library

### Integration Test Requirements
- Add integration test cases in `integration_test.go` following existing patterns
- For dynamic outputs (UUIDs, timestamps), use validation patterns in `normalizeTimestamps()`
- Test realistic use cases that demonstrate practical value

### Documentation and Examples
- Document new functions in integration tests with realistic usage examples
- Follow existing comment patterns in function implementations
- Update README.md with new function categories and examples when adding significant functionality

### Dependency Management
- Prefer standard library when possible
- For external dependencies, choose well-maintained, popular libraries
- Always run `go mod tidy` after adding dependencies
- Consider dependency size and security implications

### Testing Strategy for Dynamic Functions
- For functions with non-deterministic outputs (UUIDs, random values, timestamps):
  - Test format/pattern validity using regular expressions
  - Test functional properties (e.g., UUID v7 time ordering)
  - Use validation placeholders in integration tests (e.g., `<valid_uuid_v4>`)
  - Implement custom comparison logic in test helpers when needed

### Security Considerations
- Never implement custom cryptographic functions
- Use established libraries for security-sensitive operations
- Validate inputs to prevent injection attacks in exec-style functions
- Consider rate limiting for network-based functions

## Code Quality and Review Guidelines

### Pre-commit Checklist
- [ ] Run `go fmt ./...` to format code
- [ ] Run `go mod tidy` to clean dependencies
- [ ] Run unit tests: `go test -v ./functions`
- [ ] Run integration tests: `go test -v`
- [ ] Verify new functions work in practice with test Jsonnet files

### Branch and PR Workflow
- Create feature branches with descriptive names: `feature/function-category`
- Use clear, concise commit messages explaining the "why" not just the "what"
- Include comprehensive test coverage in PR descriptions
- Test functions manually before creating PRs

### Function Validation Criteria
- Functions should solve real-world infrastructure/configuration problems
- Avoid duplicating Jsonnet standard library functionality
- Ensure cross-platform compatibility (especially for file/path operations)
- Consider performance implications for functions that might be called frequently

## Future Enhancements
- Native functions support (currently commented out in code)
- Additional Jsonnet extensions and utility functions