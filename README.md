# jsonnet-armed

A Jsonnet rendering tool with additional useful functions.

## Installation

```bash
go install github.com/fujiwara/jsonnet-armed/cmd/jsonnet-armed@latest
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

Example Jsonnet file using external variables:
```jsonnet
{
  environment: std.extVar("env"),
  region: std.extVar("region"),
  replicas: std.extVar("replicas"),
  debug: std.extVar("debug"),
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