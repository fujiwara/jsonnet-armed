# jsonnet-armed

A Jsonnet rendering tool with additional useful functions.

## Features

jsonnet-armed provides standard Jsonnet evaluation with external variables support plus the following native functions:

#### Environment
| Function | Description | Example |
|----------|-------------|---------|
| `env(name, default)` | Get environment variable with default | [📖](#environment-functions) |
| `must_env(name)` | Get required environment variable | [📖](#environment-functions) |
| `env_parse(content)` | Parse .env format string | [📖](#environment-functions) |

#### Time
| Function | Description | Example |
|----------|-------------|---------|
| `now()` | Get current Unix timestamp | [📖](#time-functions) |
| `time_format(timestamp, format)` | Format timestamp with Go layout | [📖](#time-functions) |

#### Base64
| Function | Description | Example |
|----------|-------------|---------|
| `base64(data)` | Standard Base64 encoding | [📖](#base64-functions) |
| `base64url(data)` | URL-safe Base64 encoding | [📖](#base64-functions) |

#### Hash
| Function | Description | Example |
|----------|-------------|---------|
| `md5(data)` | MD5 hash of string | [📖](#hash-functions) |
| `sha1(data)` | SHA-1 hash of string | [📖](#hash-functions) |
| `sha256(data)` | SHA-256 hash of string | [📖](#hash-functions) |
| `sha512(data)` | SHA-512 hash of string | [📖](#hash-functions) |
| `md5_file(filename)` | MD5 hash of file content | [📖](#hash-functions) |
| `sha1_file(filename)` | SHA-1 hash of file content | [📖](#hash-functions) |
| `sha256_file(filename)` | SHA-256 hash of file content | [📖](#hash-functions) |
| `sha512_file(filename)` | SHA-512 hash of file content | [📖](#hash-functions) |

#### UUID
| Function | Description | Example |
|----------|-------------|---------|
| `uuid_v4()` | Generate random UUID v4 | [📖](#uuid-functions) |
| `uuid_v7()` | Generate time-based UUID v7 | [📖](#uuid-functions) |

#### HTTP
| Function | Description | Example |
|----------|-------------|---------|
| `http_get(url, headers)` | Make HTTP GET request | [📖](#http-functions) |
| `http_request(method, url, headers, body)` | Make HTTP request with method | [📖](#http-functions) |

#### DNS
| Function | Description | Example |
|----------|-------------|---------|
| `dns_lookup(hostname, record_type)` | DNS lookup for various record types | [📖](#dns-functions) |

#### Network
| Function | Description | Example |
|----------|-------------|---------|
| `net_port_listening(protocol, port)` | Check if a port is listening (Linux only) | [📖](#network-functions) |

#### Regular Expression
| Function | Description | Example |
|----------|-------------|---------|
| `regex_match(pattern, text)` | Check if text matches pattern | [📖](#regular-expression-functions) |
| `regex_find(pattern, text)` | Find first match | [📖](#regular-expression-functions) |
| `regex_find_all(pattern, text)` | Find all matches | [📖](#regular-expression-functions) |
| `regex_replace(pattern, replacement, text)` | Replace all matches | [📖](#regular-expression-functions) |
| `regex_split(pattern, text)` | Split text by pattern | [📖](#regular-expression-functions) |

#### JQ
| Function | Description | Example |
|----------|-------------|---------|
| `jq(query, input)` | Execute jq query on JSON data | [📖](#jq-functions) |

#### Exec
| Function | Description | Example |
|----------|-------------|---------|
| `exec(command, args)` | Execute command with arguments | [📖](#external-command-execution) |
| `exec_with_env(command, args, env)` | Execute command with custom environment | [📖](#external-command-execution) |

#### File
| Function | Description | Example |
|----------|-------------|---------|
| `file_content(filename)` | Read file content as string | [📖](#file-functions) |
| `file_stat(filename)` | Get file metadata as object | [📖](#file-functions) |
| `file_exists(filename)` | Check if file exists | [📖](#file-functions) |

#### X.509 Certificate
| Function | Description | Example |
|----------|-------------|---------|
| `x509_certificate(filename)` | Parse X.509 certificate and return detailed information | [📖](#x509-certificate-functions) |
| `x509_private_key(filename)` | Parse private key and return metadata (without exposing the key) | [📖](#x509-certificate-functions) |

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

- `-o, --output <target>`: Write output to file or HTTP(S) URL instead of stdout (can be repeated)
  - File output uses atomic writes to prevent corruption
  - HTTP(S) output sends JSON via POST request with Content-Type: application/json
  - Multiple `-o` flags can be specified to write the same output to multiple destinations
- `-S, --stdout`: Also write to stdout when using `-o/--output` (can be negated with `--no-stdout`)
- `--write-if-changed`: Write output file only if content has changed (compares using file size and SHA256 hash)
- `-V, --ext-str <key=value>`: Set external string variable (can be repeated)
- `--ext-code <key=value>`: Set external code variable (can be repeated)
- `-t, --timeout <duration>`: Timeout for evaluation (e.g., 30s, 5m, 1h)
- `-c, --cache <duration>`: Cache evaluation results for specified duration (e.g., 5m, 1h)
- `--stale <duration>`: Maximum duration to use stale cache when evaluation fails (e.g., 10m, 2h)
- `-v, --version`: Show version and exit
- `--document`: Print full documentation and exit
- `--document-toc`: Print documentation table of contents and exit
- `--document-search <keyword>`: Search documentation by keyword and print matching sections

#### Examples

Basic usage:
```bash
# Render Jsonnet to stdout
jsonnet-armed input.jsonnet

# Write output to file
jsonnet-armed -o output.json input.jsonnet

# Write to file and also display on stdout
jsonnet-armed -o output.json --stdout input.jsonnet

# Send output to HTTP endpoint
jsonnet-armed -o http://localhost:8080/webhook input.jsonnet

# Send output to HTTPS endpoint
jsonnet-armed -o https://api.example.com/config input.jsonnet

# Send to HTTP and also display on stdout
jsonnet-armed -o https://api.example.com/config -S input.jsonnet

# Write to multiple destinations simultaneously
jsonnet-armed -o output.json -o https://webhook.example.com/api input.jsonnet

# Multiple files
jsonnet-armed -o out1.json -o out2.json input.jsonnet
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

# Send output to webhook with external variables
jsonnet-armed -o https://webhook.site/unique-url -V env=prod config.jsonnet

# Use with timeout when sending to HTTP endpoint
jsonnet-armed -o http://localhost:3000/api/config -t 30s config.jsonnet
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
local http_get = std.native("http_get");
local http_request = std.native("http_request");
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

  // HTTP requests
  api_health: http_get("https://api.example.com/health", null),
  user_data: http_get("https://api.example.com/user", {
    "Authorization": "Bearer " + must_env("API_TOKEN")
  }),
  webhook_result: http_request("POST", "https://hooks.example.com/deploy", {
    "Content-Type": "application/json"
  }, std.manifestJson({
    environment: std.extVar("env"),
    commit: exec("git", ["rev-parse", "HEAD"]).stdout[0:7]
  })),
  
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

#### Adding Custom Native Functions

You can extend jsonnet-armed with your own native functions when using it as a library:

```go
package main

import (
    "bytes"
    "context"
    "fmt"

    armed "github.com/fujiwara/jsonnet-armed"
    "github.com/google/go-jsonnet"
    "github.com/google/go-jsonnet/ast"
)

func main() {
    // Create CLI instance
    cli := &armed.CLI{
        Filename: "config.jsonnet",
    }

    // Add custom native functions
    cli.AddFunctions(
        &jsonnet.NativeFunction{
            Name:   "hello",
            Params: []ast.Identifier{"name"},
            Func: func(args []any) (any, error) {
                name, ok := args[0].(string)
                if !ok {
                    return nil, fmt.Errorf("invalid argument type")
                }
                return fmt.Sprintf("Hello, %s!", name), nil
            },
        },
    )

    // Execute
    var buf bytes.Buffer
    cli.SetWriter(&buf)

    ctx := context.Background()
    if err := cli.Run(ctx); err != nil {
        panic(err)
    }

    fmt.Println(buf.String())
}
```

Your custom functions can then be used in Jsonnet files:

```jsonnet
local hello = std.native("hello");
local armed = import 'armed.libsonnet';

{
    // Using custom functions directly
    greeting: hello("World"),           // "Hello, World!"

    // Custom functions are also available in armed library
    lib_greeting: armed.hello("Armed"), // "Hello, Armed!"
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
  api_data: armed.http_get('https://api.example.com/data', null),
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

### UUID Functions
Generate UUIDs for unique identifiers.

Available UUID functions:
- `uuid_v4()`: Generate a random UUID version 4
- `uuid_v7()`: Generate a time-based UUID version 7 (sortable by timestamp)

```jsonnet
local uuid_v4 = std.native("uuid_v4");
local uuid_v7 = std.native("uuid_v7");

{
  // Generate random UUID v4
  request_id: uuid_v4(),      // e.g., "f47ac10b-58cc-4372-a567-0e02b2c3d479"
  session_id: uuid_v4(),      // e.g., "550e8400-e29b-41d4-a716-446655440000"

  // Generate time-based UUID v7 (sortable, includes timestamp)
  event_id: uuid_v7(),        // e.g., "01928b7e-7c9e-7142-9f7c-123456789abc"
  correlation_id: uuid_v7(),  // e.g., "01928b7e-7ca0-72f8-b834-987654321def"

  // Use in arrays (each call generates a new UUID)
  node_ids: [
    uuid_v4(),
    uuid_v4(),
    uuid_v4()
  ],

  // Create unique resource names
  bucket_name: "data-" + uuid_v4(),

  // Time-ordered IDs for events (v7 UUIDs are sortable by creation time)
  events: [
    {
      id: uuid_v7(),
      action: "created",
      timestamp: std.native("now")()
    },
    {
      id: uuid_v7(),
      action: "updated",
      timestamp: std.native("now")()
    }
  ]
}
```

**UUID v4 vs v7:**
- **UUID v4**: Random UUIDs, suitable for general unique identifier needs
- **UUID v7**: Time-based UUIDs that embed a timestamp, making them naturally sortable by creation time. Useful for database primary keys and event IDs where time ordering is important

### HTTP Functions

Make HTTP requests and retrieve responses directly from Jsonnet with automatic User-Agent header (`jsonnet-armed/v0.0.7`).

Available HTTP functions:
- `http_get(url, headers)`: Make GET request with optional headers
- `http_request(method, url, headers, body)`: Make HTTP request with specified method, headers, and body

Both functions return an object with:
- `status_code`: HTTP status code as number (200, 404, etc.)
- `status`: HTTP status text as string ("200 OK", "404 Not Found", etc.)
- `headers`: Response headers as object (single values as strings, multiple values as arrays)
- `body`: Response body as string

All requests have a 30-second timeout and automatically set a `User-Agent` header unless explicitly overridden.

**Error Conditions:**
HTTP functions will return an error (causing Jsonnet evaluation to fail) in the following cases:
- Invalid function arguments (non-string URL, method, or header values)
- Network connectivity issues (DNS resolution failure, connection refused)
- Request timeout (30 seconds exceeded)
- Invalid URL format or unsupported scheme
- TLS/SSL certificate validation failures (for HTTPS)

**Important:** HTTP error status codes (4xx, 5xx) do NOT cause function errors - they are returned in the response object for handling in Jsonnet.

```jsonnet
local http_get = std.native("http_get");
local http_request = std.native("http_request");

{
  // Simple GET request
  api_status: http_get("https://httpbin.org/get", null),
  // Result: {status_code: 200, status: "200 OK", headers: {...}, body: "..."}

  // GET with custom headers
  authenticated_request: http_get("https://api.example.com/user", {
    "Authorization": "Bearer your-token-here",
    "Accept": "application/json"
  }),

  // POST request with JSON body
  create_user: http_request("POST", "https://api.example.com/users", {
    "Content-Type": "application/json",
    "Authorization": "Bearer your-token-here"
  }, std.manifestJson({
    name: "John Doe",
    email: "john@example.com"
  })),

  // PUT request with custom User-Agent
  update_data: http_request("PUT", "https://api.example.com/data/123", {
    "Content-Type": "application/json",
    "User-Agent": "my-custom-client/1.0"  // Overrides default User-Agent
  }, std.manifestJson({
    updated: true,
    timestamp: std.native("now")()
  })),

  // Handle different status codes
  safe_request: {
    local response = http_get("https://api.example.com/maybe-missing", null),
    success: response.status_code == 200,
    data: if response.status_code == 200 then std.parseJson(response.body) else null,
    error: if response.status_code != 200 then response.status else null
  },

  // Error handling for network failures
  robust_request: {
    local make_request() = http_get("https://unreliable-api.example.com/data", {
      "Authorization": "Bearer token"
    }),

    // Note: Network errors cause Jsonnet evaluation to fail entirely
    // To handle this, you would need to structure your Jsonnet differently
    // or use conditional logic at a higher level
    result: make_request(),

    // For HTTP error status codes (which don't cause evaluation failure):
    is_client_error: self.result.status_code >= 400 && self.result.status_code < 500,
    is_server_error: self.result.status_code >= 500,
    success: self.result.status_code >= 200 && self.result.status_code < 300
  },

  // Working with response headers
  header_info: {
    local response = http_get("https://httpbin.org/response-headers", null),
    content_type: response.headers["Content-Type"],
    server: response.headers.Server,  // Alternative syntax
    all_headers: response.headers
  },

  // Conditional requests based on environment
  config_from_api: {
    local env = std.native("env"),
    local api_url = env("CONFIG_API_URL", "https://config.example.com"),
    local response = http_get(api_url + "/config", {
      "Authorization": "Bearer " + env("CONFIG_TOKEN", "")
    }),

    success: response.status_code == 200,
    config: if response.status_code == 200
            then std.parseJson(response.body)
            else {"default": "config"},
    error_details: if response.status_code != 200
                   then {
                     status: response.status,
                     body: response.body
                   } else null
  },

  // Webhook notification
  notify_deployment: http_request("POST", "https://hooks.slack.com/webhook/path", {
    "Content-Type": "application/json"
  }, std.manifestJson({
    text: "Deployment completed for environment: " + std.extVar("env"),
    timestamp: std.native("time_format")(std.native("now")(), "RFC3339")
  }))
}
```

**Security Notes:**
- HTTPS is recommended for all requests containing sensitive data
- Credentials should be passed via environment variables, not hardcoded
- Request and response bodies are treated as strings - parse JSON using `std.parseJson()`
- Response header names are canonicalized (Title-Case) by Go's HTTP client
- Network failures cause Jsonnet evaluation to fail - consider using cache/stale options for resilience

**Response Header Handling:**
- Single header values are returned as strings: `"application/json"`
- Multiple header values (e.g., `Set-Cookie`) are returned as arrays: `["cookie1=value1", "cookie2=value2"]`
- Header names are automatically canonicalized by Go's HTTP client (e.g., `content-type` becomes `Content-Type`, `x-custom-header` becomes `X-Custom-Header`)

### DNS Functions

Perform DNS lookups for various record types with comprehensive support for modern DNS standards.

Available DNS function:
- `dns_lookup(hostname, record_type)`: Perform DNS lookup for specified hostname and record type

The function returns an object with:
- `hostname`: The queried hostname
- `type`: The DNS record type (uppercased)
- `success`: Always true for successful lookups
- `records`: Array of DNS records (format varies by record type)

**Supported Record Types:**
- `A`: IPv4 addresses
- `AAAA`: IPv6 addresses
- `MX`: Mail exchange records with priority and hostname
- `TXT`: Text records
- `PTR`: Reverse DNS (IP to hostname)
- `CNAME`: Canonical name records
- `NS`: Name server records
- `HTTPS`: HTTPS service binding records (RFC 9460)
- `SVCB`: Service binding records (RFC 9460)

DNS lookups have a 10-second timeout by default. Network failures cause Jsonnet evaluation to fail.

```jsonnet
local dns_lookup = std.native("dns_lookup");

{
  // Basic A record lookup
  google_ips: dns_lookup("google.com", "A"),
  // Result: {hostname: "google.com", type: "A", success: true, records: ["142.250.xxx.xxx", ...]}

  // IPv6 addresses
  google_ipv6: dns_lookup("google.com", "AAAA"),
  // Result: {hostname: "google.com", type: "AAAA", success: true, records: ["2607:f8b0:xxx", ...]}

  // Mail servers
  gmail_mx: dns_lookup("gmail.com", "MX"),
  // Result: {hostname: "gmail.com", type: "MX", success: true,
  //          records: [{priority: 5, hostname: "gmail-smtp-in.l.google.com"}, ...]}

  // Text records (SPF, DKIM, etc.)
  google_txt: dns_lookup("google.com", "TXT"),
  // Result: {hostname: "google.com", type: "TXT", success: true,
  //          records: ["v=spf1 include:_spf.google.com ~all", ...]}

  // Reverse DNS lookup
  dns_google: dns_lookup("8.8.8.8", "PTR"),
  // Result: {hostname: "8.8.8.8", type: "PTR", success: true, records: ["dns.google"]}

  // Name servers
  google_ns: dns_lookup("google.com", "NS"),
  // Result: {hostname: "google.com", type: "NS", success: true,
  //          records: ["ns1.google.com", "ns2.google.com", ...]}

  // HTTPS service binding (modern HTTP/3 support detection)
  cloudflare_https: dns_lookup("cloudflare.com", "HTTPS"),
  // Result: {hostname: "cloudflare.com", type: "HTTPS", success: true,
  //          records: [{
  //            priority: 1,
  //            target: "",
  //            params: {
  //              alpn: ["h3", "h2"],
  //              ipv4hint: ["104.16.132.229", "104.16.133.229"],
  //              ipv6hint: ["2606:4700::6810:84e5", "2606:4700::6810:85e5"]
  //            }
  //          }]}

  // Case insensitive record types
  case_insensitive: dns_lookup("example.com", "a"),  // Same as "A"

  // Practical usage: Check if domain supports HTTP/3
  http3_support: {
    local https_record = dns_lookup("cloudflare.com", "HTTPS"),
    has_http3: std.length([
      record for record in https_record.records
      if std.member(std.get(std.get(record, "params", {}), "alpn", []), "h3")
    ]) > 0
  },

  // Load balancer IP discovery
  service_ips: {
    local a_records = dns_lookup("service.example.com", "A"),
    primary_ip: if std.length(a_records.records) > 0 then a_records.records[0] else null,
    all_ips: a_records.records,
    ip_count: std.length(a_records.records)
  }
}
```

**HTTPS/SVCB Record Details:**
HTTPS and SVCB records (RFC 9460) provide service binding information for modern web services:
- `priority`: Service priority (lower values preferred)
- `target`: Target hostname (empty string means same as queried hostname)
- `params`: Service parameters object containing:
  - `alpn`: Supported application protocols (e.g., `["h3", "h2"]` for HTTP/3 and HTTP/2)
  - `port`: Alternative service port (if different from default)
  - `ipv4hint`: IPv4 address hints for faster connection setup
  - `ipv6hint`: IPv6 address hints for faster connection setup

**Error Handling:**
DNS functions will return an error (causing Jsonnet evaluation to fail) in the following cases:
- Invalid function arguments (non-string hostname or record_type)
- DNS resolution failures (NXDOMAIN, timeout, server failure)
- Unsupported record types
- Network connectivity issues

### Network Functions

Check if network ports are listening on the local system by reading kernel network state.

Available network functions:
- `net_port_listening(protocol, port)`: Check if a port is listening (returns boolean)
  - `protocol`: Protocol type - "tcp", "tcp6", "udp", or "udp6" (case-insensitive)
  - `port`: Port number (1-65535)

**Platform Support:**
- **Linux only**: This function reads `/proc/net/{tcp,udp,tcp6,udp6}` to check port listening state
- **Other platforms**: Returns an error "net_port_listening is only supported on Linux"

```jsonnet
local net_port_listening = std.native("net_port_listening");

{
  // Check if common service ports are listening
  ssh_running: net_port_listening("tcp", 22),
  http_running: net_port_listening("tcp", 80),
  https_running: net_port_listening("tcp", 443),

  // Check UDP ports
  dns_running: net_port_listening("udp", 53),

  // Check IPv6 ports
  http_v6: net_port_listening("tcp6", 80),

  // Conditional configuration based on port availability
  database: {
    primary: if net_port_listening("tcp", 5432) then "localhost:5432" else "remote.db.example.com:5432",
    fallback_required: !net_port_listening("tcp", 5432)
  },

  // Service health check
  services_status: {
    nginx: net_port_listening("tcp", 80),
    postgres: net_port_listening("tcp", 5432),
    redis: net_port_listening("tcp", 6379)
  }
}
```

**Important Notes:**
- This function checks if a port is **listening**, not if it's **connectable**
- It queries the local kernel state, so it only works for ports on the same machine
- For checking remote port connectivity, use HTTP functions or implement custom checks via exec functions
- TCP ports must be in LISTEN state to return `true`
- UDP ports return `true` if the port is bound (UDP is connectionless, so there's no "LISTEN" state)
- Protocol names are case-insensitive ("TCP", "tcp", "Tcp" all work)

**Error Handling:**
The function returns an error in the following cases:
- Invalid port number (not in range 1-65535)
- Invalid protocol (not one of: tcp, tcp6, udp, udp6)
- Failed to read `/proc/net/*` files (permission denied, file not found, etc.)
- Running on non-Linux platforms

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

### Regular Expression Functions

Perform pattern matching and text manipulation using regular expressions with full Go regex syntax support.

Available regexp functions:
- `regex_match(pattern, text)`: Check if text matches the pattern (returns boolean)
- `regex_find(pattern, text)`: Find first match (returns string or null)
- `regex_find_all(pattern, text)`: Find all matches (returns array of strings)
- `regex_replace(pattern, replacement, text)`: Replace all matches (returns string)
- `regex_split(pattern, text)`: Split text by pattern (returns array of strings)

All functions use Go's `regexp` package syntax and return errors for invalid patterns. Pattern compilation is performed for each function call.

```jsonnet
local regex_match = std.native("regex_match");
local regex_find = std.native("regex_find");
local regex_find_all = std.native("regex_find_all");
local regex_replace = std.native("regex_replace");
local regex_split = std.native("regex_split");

{
  // Pattern matching for validation
  is_email: regex_match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "test@example.com"),
  is_valid_ip: regex_match("^([0-9]{1,3}\\.){3}[0-9]{1,3}$", "192.168.1.1"),
  is_uuid: regex_match("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", uuid_string),

  // Finding patterns
  first_number: regex_find("[0-9]+", "version 1.2.3"),        // "1"
  extract_domain: regex_find("@([a-zA-Z0-9.-]+)", "user@example.com"), // "@example.com"
  no_match: regex_find("xyz", "hello world"),                 // null

  // Finding all matches
  all_numbers: regex_find_all("[0-9]+", "version 1.2.3"),     // ["1", "2", "3"]
  extract_words: regex_find_all("[a-zA-Z]+", "hello 123 world 456"), // ["hello", "world"]
  ip_addresses: regex_find_all("([0-9]{1,3}\\.){3}[0-9]{1,3}", log_text),

  // Text replacement and sanitization
  sanitized_name: regex_replace("[^a-zA-Z0-9_-]", "_", "My App Name!"),  // "My_App_Name_"
  normalize_spaces: regex_replace("\\s+", " ", "hello   world    test"), // "hello world test"
  remove_tags: regex_replace("<[^>]*>", "", "<p>Hello <b>World</b></p>"), // "Hello World"

  // Environment variable substitution
  config_with_vars: regex_replace("\\$\\{([^}]+)\\}",
    std.native("env")("$1", "default"),
    "database: ${DATABASE_URL}"
  ),

  // Text splitting
  csv_parse: regex_split(",\\s*", "apple, banana, cherry"),   // ["apple", "banana", "cherry"]
  split_lines: regex_split("\\r?\\n", multi_line_text),       // Split by line breaks
  tokenize: regex_split("\\s+", "  hello   world  test  "),   // ["hello", "world", "test"]

  // Log parsing
  parse_log: {
    local log_line = "2024-01-15 10:30:45 [ERROR] Failed to connect: timeout",
    timestamp: regex_find("^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}", log_line),
    level: regex_find("\\[(\\w+)\\]", log_line),
    message: regex_replace("^[^\\]]+\\]\\s*", "", log_line)
  },

  // URL parsing
  parse_url: {
    local url = "https://api.example.com:8080/v1/users?limit=10",
    protocol: regex_find("^(https?)://", url),               // "https"
    host: regex_find("://([^:/]+)", url),                    // "api.example.com"
    port: regex_find(":([0-9]+)/", url),                     // "8080"
    path: regex_find("://[^/]+(/[^?]*)", url),               // "/v1/users"
    query: regex_find("\\?(.+)$", url)                       // "limit=10"
  },

  // Conditional logic with regex
  file_type: {
    local filename = "config.yaml",
    is_json: regex_match("\\.json$", filename),
    is_yaml: regex_match("\\.(yaml|yml)$", filename),
    is_config: regex_match("\\.(json|yaml|yml|toml|ini)$", filename)
  },

  // Version comparison
  version_info: {
    local version = "v1.2.3-alpha.1",
    major: regex_find("v?([0-9]+)", version),                // "1"
    minor: regex_find("v?[0-9]+\\.([0-9]+)", version),       // "2"
    patch: regex_find("v?[0-9]+\\.[0-9]+\\.([0-9]+)", version), // "3"
    is_prerelease: regex_match("-", version),                 // true
    prerelease_type: regex_find("-([a-zA-Z]+)", version)     // "alpha"
  }
}
```

**Pattern Syntax:**
Regular expressions use Go's RE2 syntax, which includes:
- `.`: Any character except newline
- `*`: Zero or more repetitions
- `+`: One or more repetitions
- `?`: Zero or one repetition
- `^`: Start of string
- `$`: End of string
- `[]`: Character class (e.g., `[a-z]`, `[0-9]`)
- `()`: Capturing groups
- `|`: Alternation
- `\`: Escape character

**Common Patterns:**
- Email: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- IPv4: `^([0-9]{1,3}\.){3}[0-9]{1,3}$`
- UUID: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
- Semver: `^v?([0-9]+)\.([0-9]+)\.([0-9]+)(-[a-zA-Z0-9.-]+)?$`
- URL: `^https?://[a-zA-Z0-9.-]+(/.*)?$`

**Error Handling:**
Regular expression functions will return an error (causing Jsonnet evaluation to fail) for:
- Invalid regular expression patterns
- Non-string arguments for pattern, replacement, or text parameters

### JQ Functions

Process and transform JSON data using jq query syntax with the power of the Go `gojq` library.

Available jq function:
- `jq(query, input)`: Execute jq query on input data (returns transformed data)

The jq function provides powerful JSON processing capabilities similar to the popular `jq` command-line tool:
- Field access and nested object traversal
- Array filtering, mapping, and transformation
- Complex data restructuring and aggregation
- Conditional processing and value selection

```jsonnet
local jq = std.native("jq");

{
  // Simple field access
  name: jq(".name", { name: "Alice", age: 30 }),                    // "Alice"

  // Array filtering and transformation
  adults: jq(".[] | select(.age >= 18)", [
    { name: "Alice", age: 30 },
    { name: "Bob", age: 16 },
    { name: "Charlie", age: 25 }
  ]),                                                               // [{ name: "Alice", age: 30 }, { name: "Charlie", age: 25 }]

  // Data restructuring
  summary: jq("{
    total_users: length,
    names: [.[].name],
    avg_age: (map(.age) | add / length)
  }", user_data),

  // Complex nested queries
  first_item: jq(".data.items[0].value", {
    data: { items: [{ value: "first" }, { value: "second" }] }
  }),                                                               // "first"

  // Array mapping with calculations
  doubled: jq("[.[] | . * 2]", [1, 2, 3, 4]),                    // [2, 4, 6, 8]

  // Conditional transformations
  processed: jq("map(if .status == \"active\" then .name else empty end)", status_list),

  // Grouping and aggregation
  by_category: jq("group_by(.category) | map({
    category: .[0].category,
    count: length,
    items: map(.name)
  })", items_with_categories),
}
```

The jq function returns:
- Single values for queries that produce one result
- Arrays for queries that produce multiple results
- `null` for queries that produce no results
- Structured data (objects/arrays) for complex transformations

Error handling:
- Invalid jq syntax will cause Jsonnet evaluation to fail with a descriptive error
- Non-string query arguments will return an error
- Query execution errors (e.g., accessing non-existent fields) will return an error

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

### X.509 Certificate Functions

Parse and extract information from X.509 certificates and private keys for infrastructure configuration and security validation.

Available X.509 functions:
- `x509_certificate(filename)`: Parse X.509 certificate (PEM format) and return detailed information
- `x509_private_key(filename)`: Parse private key (PEM format) and return metadata without exposing the key

**Supported Key Types:**
- RSA (PKCS#1, PKCS#8)
- ECDSA (EC PRIVATE KEY, PKCS#8)
- Ed25519 (PKCS#8)

```jsonnet
local x509_certificate = std.native("x509_certificate");
local x509_private_key = std.native("x509_private_key");

{
  // Parse certificate
  cert: x509_certificate("/etc/ssl/server.crt"),

  // Parse private key
  key: x509_private_key("/etc/ssl/server.key"),

  // Verify certificate and key pair match
  keys_match: x509_certificate("/etc/ssl/server.crt").public_key_fingerprint_sha256
           == x509_private_key("/etc/ssl/server.key").public_key_fingerprint_sha256,

  // Certificate validation
  validation: {
    local cert = x509_certificate("/etc/ssl/server.crt"),

    // Calculate days until expiration
    expires_in_days: (cert.not_after_unix - std.native("now")()) / 86400,
    is_expired: std.native("now")() > cert.not_after_unix,
    is_valid: std.native("now")() >= cert.not_before_unix
           && std.native("now")() <= cert.not_after_unix,

    // Validate domain coverage
    covers_domain: std.member(cert.dns_names, "example.com"),

    // Check if it's a CA certificate
    is_ca: cert.is_ca,
  },

  // Certificate information
  cert_info: {
    local cert = x509_certificate("/etc/ssl/server.crt"),

    // Subject information
    subject_cn: cert.subject.common_name,
    subject_org: cert.subject.organization,

    // Validity period
    valid_from: cert.not_before,
    valid_until: cert.not_after,

    // Alternative names
    domains: cert.dns_names,
    ips: cert.ip_addresses,

    // Fingerprints for verification
    fingerprint_sha256: cert.fingerprint_sha256,
    public_key_fingerprint: cert.public_key_fingerprint_sha256,

    // Key usage
    key_usage: cert.key_usage,
    ext_key_usage: cert.ext_key_usage,
  },

  // Private key information (without exposing the key itself)
  key_info: {
    local key = x509_private_key("/etc/ssl/server.key"),

    type: key.key_type,                    // "RSA", "ECDSA", "Ed25519"
    size: key.key_size,                    // 2048, 4096, etc. (RSA only)
    curve: key.curve,                      // "P-256", "P-384", etc. (ECDSA only)
    public_key_fingerprint: key.public_key_fingerprint_sha256,
  },

  // Multi-certificate setup validation
  certs: {
    local primary_cert = x509_certificate("/etc/ssl/primary.crt"),
    local backup_cert = x509_certificate("/etc/ssl/backup.crt"),
    local primary_key = x509_private_key("/etc/ssl/primary.key"),

    // Ensure primary cert and key match
    primary_valid: primary_cert.public_key_fingerprint_sha256
                == primary_key.public_key_fingerprint_sha256,

    // Check both certificates cover the same domains
    same_domains: primary_cert.dns_names == backup_cert.dns_names,

    // Select certificate based on expiration
    active_cert: if primary_cert.not_after_unix > backup_cert.not_after_unix
                  then "primary" else "backup",
  },
}
```

**Certificate Object Structure:**
```jsonnet
{
  // Fingerprints
  fingerprint_sha1: "AA:BB:CC:...",                    // SHA-1 fingerprint
  fingerprint_sha256: "11:22:33:...",                  // SHA-256 fingerprint (recommended)
  public_key_fingerprint_sha256: "55:66:77:...",       // Public key fingerprint

  // Subject and Issuer
  subject: {
    common_name: "example.com",
    organization: ["Example Corp"],
    organizational_unit: ["IT"],
    country: ["US"],
    province: ["California"],
    locality: ["San Francisco"]
  },
  issuer: { /* same structure as subject */ },

  // Validity
  serial_number: "1234567890",
  not_before: "2024-01-01T00:00:00Z",                  // RFC3339 format
  not_after: "2025-01-01T00:00:00Z",
  not_before_unix: 1704067200,                         // Unix timestamp
  not_after_unix: 1735689600,

  // Alternative Names
  dns_names: ["example.com", "*.example.com"],
  ip_addresses: ["192.0.2.1"],
  email_addresses: ["admin@example.com"],

  // Certificate Properties
  is_ca: false,
  version: 3,
  signature_algorithm: "SHA256-RSA",
  public_key_algorithm: "RSA",  // "RSA", "ECDSA", "Ed25519"

  // Usage
  key_usage: ["Digital Signature", "Key Encipherment"],
  ext_key_usage: ["Server Auth", "Client Auth"]
}
```

**Private Key Object Structure:**
```jsonnet
{
  key_type: "RSA",                                     // "RSA", "ECDSA", "Ed25519"
  key_size: 2048,                                      // RSA key size in bits
  curve: "P-256",                                      // ECDSA curve name (P-256, P-384, P-521)
  public_key_fingerprint_sha256: "55:66:77:...",      // Public key fingerprint
  public_key_pem: "-----BEGIN PUBLIC KEY-----\n..."   // Corresponding public key in PEM format
}
```

**Security Notes:**
- Private key contents are never exposed
- Only public key information and metadata are returned
- Fingerprints use industry-standard SHA-256 hashing
- Certificate and key pair matching is done via public key fingerprint comparison

**Error Handling:**
Functions return errors in the following cases:
- File not found or unreadable
- Invalid PEM format
- Unsupported certificate or key format
- Corrupted or malformed data

**Use Cases:**
- Validate certificate and private key pairs match
- Check certificate expiration dates
- Verify domain coverage in certificates
- Extract certificate fingerprints for pinning
- Automate certificate rotation checks
- Ensure multi-certificate setups are correctly configured

## Using with LLMs

jsonnet-armed embeds its full documentation into the binary. LLM agents (such as Claude Code or other AI coding assistants) can query the documentation directly from the CLI without needing access to external files or the internet.

### Available Commands

```bash
# Show table of contents to understand available features
jsonnet-armed --document-toc

# Search for specific topics (case-insensitive)
jsonnet-armed --document-search hash
jsonnet-armed --document-search "http"
jsonnet-armed --document-search dns

# Show full documentation
jsonnet-armed --document
```

### Recommended Workflow for LLM Agents

1. Run `jsonnet-armed --document-toc` to get an overview of available functions
2. Run `jsonnet-armed --document-search <keyword>` to look up specific function usage and examples
3. Use the retrieved documentation to generate correct Jsonnet code with `std.native()` calls

### Example: MCP or Tool Integration

You can configure jsonnet-armed as a tool available to your LLM agent. For example, in a system prompt or tool definition:

```
To look up jsonnet-armed native functions, run:
  jsonnet-armed --document-search <topic>

Example topics: env, hash, http, dns, exec, regex, jq, file, x509, uuid, base64, time
```

This enables the LLM to autonomously discover and correctly use jsonnet-armed's native functions when generating Jsonnet configuration files.

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
