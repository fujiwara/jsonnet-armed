# jsonnet-armed

A Jsonnet rendering tool with additional useful functions.

## Features

- Standard Jsonnet evaluation with external variables support
- Built-in native functions for environment variable access
- Hash functions for cryptographic operations

## Installation

```bash
go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed@latest
```

## Native Functions

jsonnet-armed provides built-in native functions that can be called using `std.native()`:

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

### sha256(data)
Calculate SHA256 hash of the given string and return it as hexadecimal string.

```jsonnet
local sha256 = std.native("sha256");

{
  // Returns SHA256 hash of "hello" as hex string
  hash: sha256("hello"),
  
  // Can be used with variables
  user_id: sha256(std.extVar("username")),
  
  // Combine with other functions
  short_hash: std.substr(sha256("data"), 0, 8)
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
local sha256 = std.native("sha256");

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