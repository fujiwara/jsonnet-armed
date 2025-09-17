# jsonnet-armed

A Jsonnet rendering tool with additional useful functions.

## Features

- Standard Jsonnet evaluation with external variables support
- [Environment variable access functions](#environment-functions) (`env`, `must_env`)
- [Time functions](#time-functions) for current timestamp and formatting
- [Base64 encoding functions](#base64-functions) (standard and URL-safe)
- [Hash functions](#hash-functions) for cryptographic operations
- [External command execution](#external-command-execution) with timeout and cancellation
- [File functions](#file-functions) for reading content and metadata

## Installation

```bash
go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed@latest
```

## Usage

### Command Line Usage

```bash
jsonnet-armed [options] <jsonnet-file>
```

#### Options

- `-o, --output-file <file>`: Write output to file instead of stdout (uses atomic writes to prevent corruption)
- `--write-if-changed`: Write output file only if content has changed (compares using file size and SHA256 hash)
- `-V, --ext-str <key=value>`: Set external string variable (can be repeated)
- `--ext-code <key=value>`: Set external code variable (can be repeated)
- `-t, --timeout <duration>`: Timeout for evaluation (e.g., 30s, 5m, 1h)
- `-c, --cache <duration>`: Cache evaluation results for specified duration (e.g., 5m, 1h)
- `--stale <duration>`: Maximum duration to use stale cache when evaluation fails (e.g., 10m, 2h)
- `-v, --version`: Show version and exit

#### Examples

Basic usage:
```bash
# Render Jsonnet to stdout
jsonnet-armed input.jsonnet

# Write output to file
jsonnet-armed -o output.json input.jsonnet
```

With external variables:
```bash
# Pass string variables
jsonnet-armed -V env=production -V region=us-west-2 config.jsonnet

# Pass code variables
jsonnet-armed --ext-code replicas=3 --ext-code debug=true deployment.jsonnet

# With timeout to prevent blocking operations
jsonnet-armed -t 30s config.jsonnet

# Read from stdin with timeout
echo '{ value: "test" }' | jsonnet-armed -t 10s -

# Write only if content has changed (useful for build tools)
jsonnet-armed --write-if-changed -o output.json config.jsonnet

# Cache evaluation results for 5 minutes
jsonnet-armed --cache 5m config.jsonnet

# Cache for 1 hour with external variables
jsonnet-armed --cache 1h -V env=production large-config.jsonnet

# Use stale cache fallback for reliability
jsonnet-armed --cache 5m --stale 10m config.jsonnet
```

#### Cache Feature

The cache feature stores evaluation results to avoid redundant computations:

- Cache key is generated from input file content, external variables, and output options
- Cache files are stored in `$XDG_CACHE_HOME/jsonnet-armed/` or `$HOME/.cache/jsonnet-armed/`
- Expired cache entries are automatically cleaned up
- Useful for expensive computations or frequently accessed configurations

##### Stale Cache Fallback

The `--stale` option provides resilience against evaluation failures:

- When cache expires, evaluation is attempted normally
- If evaluation fails (syntax error, missing dependencies, etc.), stale cache is used as fallback
- Stale cache is only used when fresh evaluation fails, not proactively
- Example: `--cache 5m --stale 10m` caches for 5 minutes, but allows using stale cache up to 10 minutes on errors
- Helps maintain service availability when configuration sources become temporarily unavailable

Example Jsonnet file using external variables and native functions:
```jsonnet
local env = std.native("env");
local must_env = std.native("must_env");
local md5 = std.native("md5");
local sha256 = std.native("sha256");
local sha256_file = std.native("sha256_file");
local file_content = std.native("file_content");
local file_stat = std.native("file_stat");
local exec = std.native("exec");
local exec_with_env = std.native("exec_with_env");

{
  // External variables
  environment: std.extVar("env"),
  region: std.extVar("region"),
  replicas: std.extVar("replicas"),
  debug: std.extVar("debug"),
  
  // Environment variables
  home_dir: env("HOME", "/home/user"),
  api_key: must_env("API_KEY"),
  
  // Hash functions
  config_hash: sha256(std.extVar("env") + std.extVar("region")),
  short_id: md5(std.extVar("instance_id"))[0:8],
  
  // File hash functions
  dockerfile_hash: sha256_file("Dockerfile"),
  config_file_integrity: sha256_file("/etc/app/config.yaml"),
  
  // File content and metadata
  config: std.parseJson(file_content("/etc/app/config.json")),
  config_modified: file_stat("/etc/app/config.json").mod_time,
  is_large_config: file_stat("/etc/app/config.json").size > 1024,
  
  // Command execution
  git_commit: exec("git", ["rev-parse", "HEAD"]).stdout[0:7],
  build_info: {
    local result = exec("date", ["+%Y-%m-%d %H:%M:%S"]),
    timestamp: if result.exit_code == 0 then std.strReplace(result.stdout, "\n", "") else "unknown",
    success: result.exit_code == 0
  },
  
  // System information
  system_info: {
    local uname = exec("uname", ["-a"]),
    local uptime = exec("uptime", []),
    platform: if uname.exit_code == 0 then std.strReplace(uname.stdout, "\n", "") else "unknown",
    uptime: if uptime.exit_code == 0 then std.strReplace(uptime.stdout, "\n", "") else "unknown"
  }
}
```

### Library Usage

jsonnet-armed can be embedded in your Go application as a configuration loader.

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    
    armed "github.com/fujiwara/jsonnet-armed"
)

func main() {
    // Load configuration from Jsonnet file
    config, err := LoadConfig("config.jsonnet", map[string]string{
        "env": "production",
        "region": "us-west-2",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Configuration loaded:", config)
}

func LoadConfig(filename string, vars map[string]string) (map[string]interface{}, error) {
    // Create CLI instance
    cli := &armed.CLI{
        Filename: filename,
        ExtStr:   vars,
    }
    
    // Capture output in buffer
    var buf bytes.Buffer
    cli.SetWriter(&buf)
    
    // Execute with context (supports timeout and cancellation)
    ctx := context.Background()
    if err := cli.Run(ctx); err != nil {
        return nil, err
    }
    
    // Parse JSON output
    var result map[string]interface{}
    if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
        return nil, err
    }
    
    return result, nil
}
```

You can also use timeout and external code variables:

```go
import "time"

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cli := &armed.CLI{
    Filename: "config.jsonnet",
    ExtStr: map[string]string{
        "env": "production",
    },
    ExtCode: map[string]string{
        "replicas": "3",
        "debug": "true",
    },
    Timeout: 30 * time.Second,  // Optional: CLI-level timeout
}

var buf bytes.Buffer
cli.SetWriter(&buf)

if err := cli.Run(ctx); err != nil {
    // Handle error
}
```

## Native Functions

jsonnet-armed provides built-in native functions that can be called using `std.native()`.

For convenience, you can import the `armed.libsonnet` library (dynamically generated, no separate file needed):

```jsonnet
local armed = import 'armed.libsonnet';

{
  sha256_test: armed.sha256('test'),
  env_test: armed.env('USER', 'default_user'),
  file_content: armed.file_content('config.json'),
  command_result: armed.exec('echo', ['Hello, World!']),
}
```

You can also use the traditional approach with `std.native()`:

### Environment Functions

Access environment variables with optional default values or strict requirements.

Available environment functions:
- `env(name, default)`: Get environment variable with default value
- `must_env(name)`: Get environment variable that must exist (fails if not set)
- `env_parse(content)`: Parse environment file content and return as object

```jsonnet
local env = std.native("env");
local must_env = std.native("must_env");
local env_parse = std.native("env_parse");
local file_content = std.native("file_content");

{
  // Returns the value of HOME environment variable, or "/tmp" if not set
  home: env("HOME", "/tmp"),
  
  // Can use any JSON value as default
  config: env("CONFIG", { debug: false }),
  
  // Will fail if DATABASE_URL is not set
  database_url: must_env("DATABASE_URL"),
  
  // Parse .env file content
  env_from_file: env_parse(file_content(".env")),
  
  // Parse inline env format string
  parsed_env: env_parse("KEY1=value1\nKEY2=value2\n# comment\nKEY3=value3"),
  // Result: {"KEY1": "value1", "KEY2": "value2", "KEY3": "value3"}
  
  // Use parsed env values
  local env_vars = env_parse(file_content(".env")),
  api_url: env_vars.API_URL,
  api_key: env_vars.API_KEY
}
```

### Time Functions
Work with current time and timestamp formatting.

Available time functions:
- `now()`: Get current Unix timestamp as floating-point number (nanosecond precision)
- `time_format(timestamp, format)`: Format timestamp using Go's time layout or predefined constants

**Supported Go Time Format Constants:**
For convenience, `time_format()` supports these common Go time format constant names as strings:

| Constant Name | Format String | Example Output |
|---------------|---------------|----------------|
| `"RFC3339"` | `"2006-01-02T15:04:05Z07:00"` | `"2024-01-15T10:30:45Z"` |
| `"RFC3339Nano"` | `"2006-01-02T15:04:05.999999999Z07:00"` | `"2024-01-15T10:30:45.123456789Z"` |
| `"RFC1123"` | `"Mon, 02 Jan 2006 15:04:05 MST"` | `"Mon, 15 Jan 2024 10:30:45 UTC"` |
| `"RFC1123Z"` | `"Mon, 02 Jan 2006 15:04:05 -0700"` | `"Mon, 15 Jan 2024 10:30:45 +0000"` |
| `"DateTime"` | `"2006-01-02 15:04:05"` | `"2024-01-15 10:30:45"` |
| `"DateOnly"` | `"2006-01-02"` | `"2024-01-15"` |
| `"TimeOnly"` | `"15:04:05"` | `"10:30:45"` |

```jsonnet
local now = std.native("now");
local time_format = std.native("time_format");

{
  // Current time
  timestamp: now(),                        // 1705314645.123456789
  
  // Using predefined format constants (recommended)
  iso_time: time_format(now(), "RFC3339"),             // "2024-01-15T10:30:45Z"
  date_only: time_format(now(), "DateOnly"),           // "2024-01-15"
  time_only: time_format(now(), "TimeOnly"),           // "10:30:45" 
  readable: time_format(now(), "DateTime"),            // "2024-01-15 10:30:45"
  
  // Using custom Go time format strings  
  custom: time_format(now(), "2006/01/02 15:04:05"),   // "2024/01/15 10:30:45"
  year_month: time_format(now(), "2006-01"),           // "2024-01"
  
  // Format specific timestamp
  formatted: time_format(1705314645.123456789, "RFC3339Nano"), // "2024-01-15T10:30:45.123456716Z"
  
  // Useful for TTL, expiration times
  timestamp_ms: now() * 1000,              // Convert to milliseconds
  timestamp_sec: std.floor(now()),         // Integer seconds only
}
```

### Base64 Functions
Encode strings to Base64 format.

Available base64 functions:
- `base64(data)`: Standard Base64 encoding
- `base64url(data)`: URL-safe Base64 encoding (uses `-` and `_` instead of `+` and `/`)

```jsonnet
local base64 = std.native("base64");
local base64url = std.native("base64url");

{
  // Standard Base64 encoding
  encoded: base64("Hello, World!"),        // "SGVsbG8sIFdvcmxkIQ=="
  empty: base64(""),                       // ""
  
  // URL-safe Base64 encoding  
  url_safe: base64url("??>>"),             // "Pz8-Pg==" (uses - instead of +)
  
  // Encoding JSON data
  json_encoded: base64(std.manifestJsonEx({ 
    user: "admin", 
    timestamp: 1234567890 
  }, "")),
  
  // Encoding with special characters
  unicode: base64("こんにちは世界"),        // "44GT44KT44Gr44Gh44Gv5LiW55WM"
}
```

### Hash Functions
Calculate hash of the given string or file and return it as hexadecimal string.

Available hash functions:

**String Hash Functions:**
- `md5(data)`: MD5 hash (32 characters)
- `sha1(data)`: SHA-1 hash (40 characters)  
- `sha256(data)`: SHA-256 hash (64 characters)
- `sha512(data)`: SHA-512 hash (128 characters)

**File Hash Functions:**
- `md5_file(filename)`: MD5 hash of file content (32 characters)
- `sha1_file(filename)`: SHA-1 hash of file content (40 characters)
- `sha256_file(filename)`: SHA-256 hash of file content (64 characters)
- `sha512_file(filename)`: SHA-512 hash of file content (128 characters)

```jsonnet
local md5 = std.native("md5");
local sha1 = std.native("sha1");
local sha256 = std.native("sha256");
local sha512 = std.native("sha512");

local md5_file = std.native("md5_file");
local sha256_file = std.native("sha256_file");

{
  // String hash functions
  md5_hash: md5("hello"),       // "5d41402abc4b2a76b9719d911017c592"
  sha1_hash: sha1("hello"),     // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
  sha256_hash: sha256("hello"), // "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
  sha512_hash: sha512("hello"), // 128 character hash
  
  // File hash functions
  config_file_hash: sha256_file("/etc/config.json"),
  self_hash: md5_file(std.thisFile),
  
  // Compare string vs file content
  matches: sha256("hello") == sha256_file("/tmp/hello.txt"),
  
  // Can be used with variables
  user_id: sha256(std.extVar("username")),
  
  // Combine with other functions
  short_hash: std.substr(sha256("data"), 0, 8)
}
```

### External Command Execution

Execute external commands and capture their output, with timeout and cancellation support.

Available exec functions:
- `exec(command, args)`: Execute command with arguments array
- `exec_with_env(command, args, env_vars)`: Execute command with custom environment variables

Both functions return an object with:
- `stdout`: Standard output as string
- `stderr`: Standard error as string  
- `exit_code`: Exit code as number (0 = success)

Commands are executed with a 30-second timeout by default (configurable via `functions.DefaultExecTimeout`). When the CLI timeout is reached, running commands are cancelled immediately.

```jsonnet
local exec = std.native("exec");
local exec_with_env = std.native("exec_with_env");

{
  // Basic command execution
  hello: exec("echo", ["Hello, World!"]),
  // Result: {stdout: "Hello, World!\n", stderr: "", exit_code: 0}
  
  // Command with multiple arguments
  ls_result: exec("ls", ["-la", "/tmp"]),
  
  // Check exit code for success
  success: exec("true", []).exit_code == 0,   // true
  failure: exec("false", []).exit_code == 0,  // false
  
  // Capture stderr
  error_output: exec("sh", ["-c", "echo error >&2; exit 1"]),
  // Result: {stdout: "", stderr: "error\n", exit_code: 1}
  
  // Command with environment variables
  custom_env: exec_with_env("sh", ["-c", "echo $CUSTOM_VAR"], {
    "CUSTOM_VAR": "Hello from env!"
  }),
  // Result: {stdout: "Hello from env!\n", stderr: "", exit_code: 0}
  
  // Git commands with clean environment
  git_status: exec_with_env("git", ["status", "--porcelain"], {
    "GIT_CONFIG_NOGLOBAL": "1",
    "GIT_CONFIG_NOSYSTEM": "1"
  }),
  
  // Conditional execution based on exit code
  file_info: {
    local test_result = exec("test", ["-f", "/etc/passwd"]),
    file_exists: test_result.exit_code == 0,
    content: if test_result.exit_code == 0 
             then exec("head", ["-1", "/etc/passwd"]).stdout
             else "File not found"
  },
  
  // Safe command execution with error handling
  safe_command: {
    local result = exec("unknown-command", ["arg1"]),
    success: result.exit_code == 0,
    output: if result.exit_code == 0 then result.stdout else result.stderr,
    error_msg: if result.exit_code != 0 then "Command failed: " + result.stderr else null
  },
  
  // Working with JSON output
  docker_info: {
    local result = exec("docker", ["inspect", "container-name"]),
    success: result.exit_code == 0,
    data: if result.exit_code == 0 then std.parseJson(result.stdout) else null
  }
}
```

**Security Notes:**
- Commands are executed directly (no shell interpretation)
- Arguments are properly isolated to prevent command injection
- Use environment variables via `exec_with_env` rather than shell expansion
- Commands timeout after 30 seconds and are forcefully killed (SIGTERM then SIGKILL)

**Timeout Behavior:**
- Default timeout: 30 seconds (configurable via `functions.DefaultExecTimeout`)
- If CLI has `--timeout` flag, exec commands are cancelled when CLI times out
- Process termination: SIGTERM → 5 second grace period → SIGKILL

### File Functions
Access file content and metadata directly from Jsonnet.

Available file functions:
- `file_content(filename)`: Read file content as string
- `file_stat(filename)`: Get file metadata as object
- `file_exists(filename)`: Check if file or directory exists (returns boolean)

```jsonnet
local file_content = std.native("file_content");
local file_stat = std.native("file_stat");
local file_exists = std.native("file_exists");

{
  // Check file existence
  config_exists: file_exists("/etc/config.json"),        // true/false
  readme_exists: file_exists("README.md"),               // true/false
  dir_exists: file_exists("/etc"),                       // true (works for directories too)
  
  // Conditional file reading
  config: if file_exists("/etc/config.json") 
          then std.parseJson(file_content("/etc/config.json"))
          else {"default": "config"},
  
  // Read file content (only if exists)
  config_content: if file_exists("/etc/config.json") 
                  then file_content("/etc/config.json") 
                  else "File not found",
  
  // Get file metadata (only if exists)
  config_stat: if file_exists("/etc/config.json") 
               then file_stat("/etc/config.json")
               else null,
  
  // File information includes:
  // - name: filename
  // - size: file size in bytes
  // - mode: file permissions as string
  // - mod_time: modification time as Unix timestamp
  // - is_dir: true if directory, false if regular file
  
  // Safe file operations
  file_info: {
    local filename = "data.txt",
    exists: file_exists(filename),
    content: if file_exists(filename) then file_content(filename) else null,
    stat: if file_exists(filename) then file_stat(filename) else null,
    safe_size: if file_exists(filename) then file_stat(filename).size else 0
  }
}
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/fujiwara/jsonnet-armed.git
cd jsonnet-armed

# Build
make

# Run tests
make test

# Install
make install
```

## Requirements

- Go 1.24 or later

## License

MIT

## Author

Fujiwara Shunichiro