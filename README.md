# jsonnet-armed

A Jsonnet rendering tool with additional useful functions.

## Features

- Standard Jsonnet evaluation with external variables support
- Built-in native functions for environment variable access
- Hash functions for cryptographic operations
- File functions for reading content and metadata

## Installation

```bash
go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed@latest
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
}
```

You can also use the traditional approach with `std.native()`:

### env(name, default)
Get an environment variable with a default value.

```jsonnet
local env = std.native("env");

{
  // Returns the value of HOME environment variable, or "/tmp" if not set
  home: env("HOME", "/tmp"),
  
  // Can use any JSON value as default
  config: env("CONFIG", { debug: false })
}
```

### must_env(name)
Get an environment variable that must exist. Returns an error if the variable is not set.

```jsonnet
local must_env = std.native("must_env");

{
  // Will fail if DATABASE_URL is not set
  database_url: must_env("DATABASE_URL")
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

### File Functions
Access file content and metadata directly from Jsonnet.

Available file functions:
- `file_content(filename)`: Read file content as string
- `file_stat(filename)`: Get file metadata as object

```jsonnet
local file_content = std.native("file_content");
local file_stat = std.native("file_stat");

{
  // Read file content
  config_content: file_content("/etc/config.json"),
  readme: file_content("README.md"),
  
  // Parse JSON files directly
  config: std.parseJson(file_content("/etc/config.json")),
  
  // Get file metadata
  config_stat: file_stat("/etc/config.json"),
  
  // File information includes:
  // - name: filename
  // - size: file size in bytes
  // - mode: file permissions as string
  // - mod_time: modification time as Unix timestamp
  // - is_dir: true if directory, false if regular file
  
  // Combine content and metadata
  file_info: {
    content: file_content("data.txt"),
    stat: file_stat("data.txt"),
    content_matches_size: std.length(file_content("data.txt")) == file_stat("data.txt").size
  }
}
```

## Usage

```bash
jsonnet-armed [options] <jsonnet-file>
```

### Options

- `-o, --output-file <file>`: Write output to file instead of stdout
- `-V, --ext-str <key=value>`: Set external string variable (can be repeated)
- `--ext-code <key=value>`: Set external code variable (can be repeated)
- `-v, --version`: Show version and exit

### Examples

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
```

Example Jsonnet file using external variables and native functions:
```jsonnet
local env = std.native("env");
local must_env = std.native("must_env");
local md5 = std.native("md5");
local sha256 = std.native("sha256");
local sha256_file = std.native("sha256_file");
local file_content = std.native("file_content");
local file_stat = std.native("file_stat");

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