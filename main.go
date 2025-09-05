package armed

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fujiwara/jsonnet-armed/functions"
	"github.com/google/go-jsonnet"
)

var output io.Writer = os.Stdout

// SetOutput sets the output destination for jsonnet evaluation results
func SetOutput(w io.Writer) {
	output = w
}

func Run(ctx context.Context) error {
	cli := &CLI{}
	kong.Parse(cli, kong.Vars{"version": fmt.Sprintf("jsonnet-armed %s", Version)})
	return run(ctx, cli)
}

// RunWithCLI runs the jsonnet evaluation with the given CLI configuration
func RunWithCLI(ctx context.Context, cli *CLI) error {
	return run(ctx, cli)
}

func run(ctx context.Context, cli *CLI) error {
	vm := jsonnet.MakeVM()

	// Add importer for armed.libsonnet
	vm.Importer(&ArmedImporter{})

	// Register native functions
	for _, f := range functions.AllFunctions() {
		vm.NativeFunction(f)
	}

	for k, v := range cli.ExtStr {
		vm.ExtVar(k, v)
	}
	for k, v := range cli.ExtCode {
		vm.ExtCode(k, v)
	}

	var jsonStr string
	var err error

	if cli.Filename == "-" {
		// Read from stdin
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		jsonStr, err = vm.EvaluateAnonymousSnippet("stdin", string(input))
	} else {
		jsonStr, err = vm.EvaluateFile(cli.Filename)
	}
	if err != nil {
		return fmt.Errorf("failed to evaluate: %w", err)
	}

	if cli.OutputFile != "" {
		return os.WriteFile(cli.OutputFile, []byte(jsonStr), 0644)
	}
	_, err = io.WriteString(output, jsonStr)
	return err
}

// ArmedImporter provides virtual file system for armed.libsonnet
type ArmedImporter struct{}

func (ai *ArmedImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	if importedPath == "armed.libsonnet" {
		// Generate the library content dynamically
		content := functions.GenerateArmedLib()
		return jsonnet.MakeContents(content), "armed.libsonnet", nil
	}

	// Fall back to default file system import
	importer := &jsonnet.FileImporter{}
	return importer.Import(importedFrom, importedPath)
}
